import { useMemo, useState, useRef, useEffect, type ChangeEvent } from 'react';
import { Layout, Menu, Grid, Button, Message, Modal, Drawer, Input, Spin } from '@arco-design/web-react';
import { IconExport, IconImport, IconMenu } from '@arco-design/web-react/icon';
import SubscribeManage from './pages/SubscribeManage';
import NodeGroupManage from './pages/NodeGroupManage';
import RuleSetManage from './pages/RuleSetManage';
import SettingManage from './pages/SettingManage';
import DeviceManage from './pages/DeviceManage';
import InboundManage from './pages/InboundManage';
import WireGuardManage from './pages/WireGuardManage';
import OutboundManage from './pages/OutboundManage';
import DnsManage from './pages/DnsManage';
import { MANAGEMENT_NAV_ITEMS } from './utils/navigation';
import * as api from './api';
import { formatConfigImportSummary } from './utils/configTransfer';
import type { AuthProfile } from './types';
import './App.css';

const Sider = Layout.Sider;
const Content = Layout.Content;
const Row = Grid.Row;
const Col = Grid.Col;

function getDownloadFilename(contentDisposition?: string) {
  if (!contentDisposition) {
    return `singboxconfig-export-${new Date().toISOString().replace(/[-:]/g, '').replace(/\..+/, '').replace('T', '-')}.json`;
  }
  const match = contentDisposition.match(/filename="?([^"]+)"?/i);
  return match?.[1] || `singboxconfig-export-${Date.now()}.json`;
}

function App() {
  const [activeKey, setActiveKey] = useState('subscribes');
  const [exporting, setExporting] = useState(false);
  const [importing, setImporting] = useState(false);
  const [isMobile, setIsMobile] = useState(window.innerWidth <= 768);
  const [menuVisible, setMenuVisible] = useState(false);
  const [authProfile, setAuthProfile] = useState<AuthProfile | null>(null);
  const [authReady, setAuthReady] = useState(false);
  const [authToken, setAuthToken] = useState<string | null>(api.getStoredAuthToken());
  const [loggingIn, setLoggingIn] = useState(false);
  const [loginForm, setLoginForm] = useState({ username: '', password: '' });
  const [accountModalVisible, setAccountModalVisible] = useState(false);
  const [updatingAccount, setUpdatingAccount] = useState(false);
  const [accountForm, setAccountForm] = useState({
    oldPassword: '',
    newUsername: '',
    newPassword: '',
    confirmPassword: '',
  });
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  useEffect(() => {
    const handleResize = () => setIsMobile(window.innerWidth <= 768);
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  const loadAuthProfile = async () => {
    if (!api.getStoredAuthToken()) {
      setAuthProfile(null);
      setAuthReady(true);
      return false;
    }
    try {
      const response = await api.getAuthMe();
      setAuthProfile(response.data);
      setAuthReady(true);
      return true;
    } catch {
      api.clearStoredAuthToken();
      setAuthToken(null);
      setAuthProfile(null);
      setAuthReady(true);
      return false;
    }
  };

  useEffect(() => {
    void loadAuthProfile();
  }, []);

  const activeItem = useMemo(
    () => MANAGEMENT_NAV_ITEMS.find((item) => item.key === activeKey) ?? MANAGEMENT_NAV_ITEMS[0],
    [activeKey],
  );

  const resetAccountForm = () => {
    setAccountForm({
      oldPassword: '',
      newUsername: authProfile?.username || '',
      newPassword: '',
      confirmPassword: '',
    });
  };

  const openAccountModal = () => {
    resetAccountForm();
    setAccountModalVisible(true);
  };

  const closeAccountModal = () => {
    setUpdatingAccount(false);
    setAccountModalVisible(false);
    resetAccountForm();
  };

  const handleLogin = async () => {
    if (!loginForm.username || !loginForm.password) {
      Message.error('请输入用户名和密码');
      return;
    }
    try {
      setLoggingIn(true);
      const response = await api.login(loginForm.username.trim(), loginForm.password);
      api.setStoredAuthToken(response.data.access_token);
      setAuthToken(response.data.access_token);
      setLoginForm({ username: '', password: '' });
      await loadAuthProfile();
      Message.success(`已登录为 ${response.data.username}`);
    } catch (error: any) {
      Message.error(error?.response?.data?.error || '登录失败');
    } finally {
      setLoggingIn(false);
    }
  };

  const handleLogout = () => {
    api.clearStoredAuthToken();
    setAuthToken(null);
    setAuthProfile(null);
    setAccountModalVisible(false);
    setMenuVisible(false);
    Message.info('已退出登录');
  };

  const handleExportConfig = async () => {
    try {
      setExporting(true);
      const response = await api.exportConfig();
      const blob = new Blob([response.data], { type: 'application/json;charset=utf-8' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = getDownloadFilename(response.headers['content-disposition']);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
      Message.success('配置导出成功');
    } catch {
      Message.error('配置导出失败');
    } finally {
      setExporting(false);
    }
  };

  const submitImport = async (file: File) => {
    try {
      setImporting(true);
      const response = await api.importConfig(file);
      Message.success('导入完成');
      Modal.info({
        title: '导入结果',
        content: (
          <pre style={{ margin: 0, whiteSpace: 'pre-wrap', wordBreak: 'break-word', fontSize: 12 }}>
            {formatConfigImportSummary(response.data)}
          </pre>
        ),
      });
    } catch {
      Message.error('导入失败');
    } finally {
      setImporting(false);
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    }
  };

  const handleImportFileChange = (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;
    if (!file.name.toLowerCase().endsWith('.json')) {
      Message.error('请选择 .json 文件');
      event.target.value = '';
      return;
    }
    Modal.confirm({
      title: '确认导入配置',
      content: '导入文件将与当前数据合并。已存在的资源将跳过。',
      onOk: () => submitImport(file),
      onCancel: () => { event.target.value = ''; },
    });
  };

  const handleUpdateAccount = async () => {
    const nextUsername = accountForm.newUsername.trim();
    const passwordChanged = Boolean(accountForm.newPassword);
    const usernameChanged = nextUsername !== '' && nextUsername !== authProfile?.username;

    if (!accountForm.oldPassword) {
      Message.error('请输入当前密码');
      return;
    }
    if (!usernameChanged && !passwordChanged) {
      Message.error('请至少修改用户名或密码中的一项');
      return;
    }
    if (passwordChanged && accountForm.newPassword !== accountForm.confirmPassword) {
      Message.error('两次输入的新密码不一致');
      return;
    }

    try {
      setUpdatingAccount(true);
      const response = await api.changeCredentials(
        accountForm.oldPassword,
        usernameChanged ? nextUsername : undefined,
        passwordChanged ? accountForm.newPassword : undefined,
      );
      api.setStoredAuthToken(response.data.access_token);
      setAuthToken(response.data.access_token);
      await loadAuthProfile();
      closeAccountModal();
      Message.success('账号信息已更新');
    } catch (error: any) {
      Message.error(error?.response?.data?.error || '账号更新失败');
      setUpdatingAccount(false);
    }
  };

  const renderMenu = () => (
    <Menu
      className="app-menu"
      selectedKeys={[activeKey]}
      onClickMenuItem={(key) => {
        setActiveKey(key);
        setMenuVisible(false);
      }}
    >
      {MANAGEMENT_NAV_ITEMS.map((item) => (
        <Menu.Item key={item.key}>{item.title}</Menu.Item>
      ))}
    </Menu>
  );

  const renderPage = () => {
    switch (activeKey) {
      case 'subscribes': return <SubscribeManage />;
      case 'node-groups': return <NodeGroupManage />;
      case 'rule-sets': return <RuleSetManage />;
      case 'settings': return <SettingManage />;
      case 'devices': return <DeviceManage />;
      case 'inbounds': return <InboundManage />;
      case 'wireguards': return <WireGuardManage />;
      case 'outbounds': return <OutboundManage />;
      case 'dns': return <DnsManage />;
      default: return <SubscribeManage />;
    }
  };

  if (!authReady) {
    return (
      <div className="auth-screen">
        <div className="auth-panel auth-panel-loading">
          <Spin size={28} />
          <div className="auth-panel-title">正在验证登录态</div>
        </div>
      </div>
    );
  }

  if (!authToken || !authProfile) {
    return (
      <div className="auth-screen">
        <div className="auth-panel">
          <div className="auth-panel-eyebrow">SingBox Config</div>
          <div className="auth-panel-title">管理员登录</div>
          <div className="auth-panel-description"></div>
          <div className="auth-form-field">
            <div className="auth-form-label">用户名</div>
            <Input
              value={loginForm.username}
              onChange={(value) => setLoginForm((prev) => ({ ...prev, username: value }))}
              autoComplete="username"
              placeholder="请输入管理员用户名"
              onPressEnter={handleLogin}
            />
          </div>
          <div className="auth-form-field">
            <div className="auth-form-label">密码</div>
            <Input.Password
              value={loginForm.password}
              onChange={(value) => setLoginForm((prev) => ({ ...prev, password: value }))}
              autoComplete="current-password"
              placeholder="请输入管理员密码"
              onPressEnter={handleLogin}
            />
          </div>
          <Button type="primary" long loading={loggingIn} onClick={handleLogin} style={{ marginTop: 20 }}>
            登录后台
          </Button>
          <div className="auth-panel-hint">如果你忘记密码，仍可使用服务端的 `-reset-password` 或 `FORCE_RESET_PASSWORD` 重置。</div>
        </div>
      </div>
    );
  }

  return (
    <Layout className="app-shell">
      {isMobile ? (
        <>
          <div className="mobile-header">
            <Button
              shape="circle"
              type="text"
              icon={<IconMenu />}
              onClick={() => setMenuVisible(true)}
              style={{ fontSize: 20, color: 'var(--text-primary)' }}
            />
            <div className="mobile-brand-title">SingBox 控制中心</div>
            <div style={{ width: 32 }} />
          </div>
          <Drawer
            visible={menuVisible}
            placement="left"
            onCancel={() => setMenuVisible(false)}
            footer={null}
            headerStyle={{ display: 'none' }}
            width={240}
            className="mobile-menu-drawer"
          >
            <div className="app-brand">
              <div className="app-brand-title">SingBox</div>
              <div className="app-brand-subtitle">控制中心</div>
            </div>
            {renderMenu()}
          </Drawer>
        </>
      ) : (
        <Sider width={200} className="app-sider" breakpoint="lg" collapsedWidth={0}>
          <div className="app-brand">
            <div className="app-brand-title">SingBox</div>
            <div className="app-brand-subtitle">控制中心</div>
          </div>
          {renderMenu()}
        </Sider>
      )}

      <Content className="app-content">
        <Row justify="center">
          <Col span={24}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 24, flexWrap: 'wrap', gap: 12 }}>
              <div className="page-header" style={{ marginBottom: 0 }}>
                <div className="page-title">{activeItem.title}</div>
                <div className="page-description">{activeItem.description}</div>
              </div>
              <div className="header-actions">
                <Button className="auth-pill" onClick={openAccountModal}>
                  <span className="auth-pill-label">管理员</span>
                  <span className="auth-pill-value">{authProfile.username}</span>
                </Button>
                <Button
                  shape="circle"
                  icon={<IconExport />}
                  loading={exporting}
                  onClick={handleExportConfig}
                  title="导出 JSON"
                  style={{ background: 'rgba(255,255,255,0.6)', border: '1px solid var(--border-color)' }}
                />
                <Button
                  shape="circle"
                  icon={<IconImport />}
                  loading={importing}
                  onClick={() => fileInputRef.current?.click()}
                  title="导入 JSON"
                  style={{ background: 'rgba(255,255,255,0.6)', border: '1px solid var(--border-color)' }}
                />
                <Button onClick={handleLogout}>退出</Button>

                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".json,application/json"
                  style={{ display: 'none' }}
                  onChange={handleImportFileChange}
                />
              </div>
            </div>
            <div className="page-main-container">
              {renderPage()}
            </div>
          </Col>
        </Row>
      </Content>
      <Modal
        visible={accountModalVisible}
        title="账号设置"
        confirmLoading={updatingAccount}
        onOk={handleUpdateAccount}
        onCancel={closeAccountModal}
      >
        <div className="auth-modal-note">
          当前账号：<strong>{authProfile.username}</strong>
        </div>
        <div className="auth-modal-note">
          修改用户名或密码后，当前登录态会刷新为新的 Bearer Token，旧 token 会立即失效。
        </div>
        <div className="auth-form-field">
          <div className="auth-form-label">新用户名</div>
          <Input
            value={accountForm.newUsername}
            onChange={(value) => setAccountForm((prev) => ({ ...prev, newUsername: value }))}
            autoComplete="username"
            placeholder="留空表示保持当前用户名"
          />
        </div>
        <div className="auth-form-field">
          <div className="auth-form-label">当前密码</div>
          <Input.Password
            value={accountForm.oldPassword}
            onChange={(value) => setAccountForm((prev) => ({ ...prev, oldPassword: value }))}
            autoComplete="current-password"
            placeholder="请输入当前密码确认变更"
          />
        </div>
        <div className="auth-form-field">
          <div className="auth-form-label">新密码</div>
          <Input.Password
            value={accountForm.newPassword}
            onChange={(value) => setAccountForm((prev) => ({ ...prev, newPassword: value }))}
            autoComplete="new-password"
            placeholder="至少 8 位，留空表示不修改密码"
          />
        </div>
        <div className="auth-form-field">
          <div className="auth-form-label">确认新密码</div>
          <Input.Password
            value={accountForm.confirmPassword}
            onChange={(value) => setAccountForm((prev) => ({ ...prev, confirmPassword: value }))}
            autoComplete="new-password"
            placeholder="如修改密码，请再次输入"
          />
        </div>
      </Modal>
    </Layout>
  );
}

export default App;
