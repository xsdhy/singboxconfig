import { useState, useEffect, useCallback } from 'react';
import { Message, Input, Button, Space } from '@arco-design/web-react';
import SettingTable from '../components/SettingTable';
import SettingModal from '../components/SettingModal';
import PageToolbar from '../components/PageToolbar';
import DataState from '../components/DataState';
import * as api from '../api';
import type { Setting } from '../types';
import { SYSTEM_HOST_SETTING_KEY, isValidSystemHost, normalizeSystemHost } from '../utils/systemHost';

export default function SettingManage() {
  const [settings, setSettings] = useState<Setting[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [deletingKey, setDeletingKey] = useState<string | null>(null);
  const [setVisible, setSetVisible] = useState(false);
  const [setTitle, setSetTitle] = useState('添加设置');
  const [setForm, setSetForm] = useState<Partial<Setting>>({});

  // 系统 Host 独立编辑入口状态。
  const [systemHost, setSystemHost] = useState('');
  const [systemHostSaving, setSystemHostSaving] = useState(false);

  const loadSystemHost = useCallback(async () => {
    try {
      const res = await api.getSettingByKey(SYSTEM_HOST_SETTING_KEY);
      setSystemHost(res.data.value || '');
    } catch {
      // 未配置时接口返回 404，按空值处理。
      setSystemHost('');
    }
  }, []);

  const loadData = useCallback(async (manual = false) => {
    try {
      if (manual) {
        setRefreshing(true);
      } else {
        setLoading(true);
      }
      const res = await api.getSettings();
      setSettings(Array.isArray(res.data) ? res.data : []);
    } catch {
      Message.error('加载设置数据失败');
      setSettings([]);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => { loadData(); }, [loadData]);
  useEffect(() => { loadSystemHost(); }, [loadSystemHost]);

  // 保存系统 Host：先做前端合法性校验，再写入全局设置（不存在则创建）。
  const handleSaveSystemHost = async () => {
    const normalized = normalizeSystemHost(systemHost);
    if (!isValidSystemHost(normalized)) {
      Message.error('系统 Host 必须是合法的 http(s) 绝对地址，例如 https://config.example.com');
      return;
    }
    try {
      setSystemHostSaving(true);
      const payload = { key: SYSTEM_HOST_SETTING_KEY, value: normalized };
      try {
        await api.updateSetting(SYSTEM_HOST_SETTING_KEY, payload);
      } catch {
        await api.createSetting(payload);
      }
      setSystemHost(normalized);
      Message.success('系统 Host 保存成功');
      await loadData();
    } catch {
      Message.error('系统 Host 保存失败');
    } finally {
      setSystemHostSaving(false);
    }
  };

  const handleAddSet = () => { setSetForm({}); setSetTitle('添加设置'); setSetVisible(true); };
  const handleEditSet = (r: Setting) => { setSetForm({ ...r }); setSetTitle('编辑设置'); setSetVisible(true); };
  const handleSetOk = async (values: Setting) => {
    try {
      setSubmitting(true);
      if (setForm.key) {
        await api.updateSetting(setForm.key, { ...setForm, ...values } as Setting);
      } else {
        await api.createSetting(values);
      }
      Message.success('保存成功');
      setSetVisible(false);
      await loadData();
    } catch { Message.error('保存失败'); }
    finally { setSubmitting(false); }
  };
  const handleDeleteSet = async (r: Setting) => {
    try {
      setDeletingKey(r.key);
      await api.deleteSetting(r.key);
      Message.success('删除成功');
      await loadData();
    } catch { Message.error('删除失败'); }
    finally { setDeletingKey(null); }
  };

  return (
    <>
      <PageToolbar
        summary="全局设置：配置系统全局参数。"
        count={settings.length}
        countLabel="项设置"
        onRefresh={() => loadData(true)}
        refreshing={refreshing}
        onPrimaryAction={handleAddSet}
        primaryActionLabel="添加设置"
      />

      <div style={{ paddingTop: 8 }}>
        <div
          style={{
            border: '1px solid var(--border-color)',
            borderRadius: 12,
            padding: 16,
            marginBottom: 16,
            background: 'rgba(255,255,255,0.6)',
          }}
        >
          <div style={{ fontWeight: 600, marginBottom: 4 }}>系统 Host</div>
          <div style={{ fontSize: 12, color: 'var(--text-secondary)', marginBottom: 12 }}>
            本服务对外可访问的基础地址（如 https://config.example.com，不带尾斜杠）。
            配置后，生成的整份配置会把本地规则集改为指向本服务的远程规则集 URL；留空则规则集仍按内联/展开方式生成。
          </div>
          <Space>
            <Input
              style={{ width: 360 }}
              value={systemHost}
              onChange={setSystemHost}
              placeholder="https://config.example.com"
              allowClear
            />
            <Button type="primary" loading={systemHostSaving} onClick={handleSaveSystemHost}>
              保存系统 Host
            </Button>
          </Space>
        </div>

        <DataState
          loading={loading}
          isEmpty={settings.length === 0}
          emptyTitle="还没有全局设置"
          emptyDescription="全局设置会影响多个功能模块，建议先录入系统级公共参数。"
          createLabel="立即添加设置"
          onCreate={handleAddSet}
        >
          <SettingTable
            data={settings}
            deletingKey={deletingKey}
            onEdit={handleEditSet}
            onDelete={handleDeleteSet}
          />
        </DataState>
      </div>

      <SettingModal visible={setVisible} title={setTitle} initialValues={setForm}
        confirmLoading={submitting}
        onOk={handleSetOk} onCancel={() => setSetVisible(false)} />
    </>
  );
}
