import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Button,
  Checkbox,
  Drawer,
  Form,
  Grid,
  Input,
  InputNumber,
  Message,
  Modal,
  Space,
  Switch,
  Tag,
  Typography,
} from '@arco-design/web-react';
import * as api from '../api';
import DataState from '../components/DataState';
import PageToolbar from '../components/PageToolbar';
import type { Device, Inbound, WireGuard } from '../types';
import { buildDeleteConfirmContent } from '../utils/deleteConfirm';
import { buildDeviceInboundPayload, buildDeviceInboundSelection } from '../utils/deviceManagement';

const FormItem = Form.Item;
const TextArea = Input.TextArea;
const Row = Grid.Row;
const Col = Grid.Col;
const { Title, Text } = Typography;

// getWireGuardLabel 用于把设备上的 wireGuardTag 转成更友好的名称展示。
function getWireGuardLabel(record: Device, wireGuards: WireGuard[]) {
  if (!record.wireGuardTag) {
    return '未绑定';
  }
  const matched = wireGuards.find((item) => item.tag === record.wireGuardTag);
  return matched ? `${matched.name} (${matched.tag})` : record.wireGuardTag;
}

// getOpenAPIOrigin 生成公开配置链接的源地址。
// Vite 开发环境没有代理 /open/*，因此本地 5173 页面复制链接时指向后端默认端口。
function getOpenAPIOrigin() {
  if (window.location.hostname === 'localhost' && window.location.port === '5173') {
    return 'http://localhost:7391';
  }
  return window.location.origin;
}

// buildDeviceConfigURL 根据设备和输出格式构造可直接给客户端订阅的公开链接。
function buildDeviceConfigURL(record: Device, outputType: 'singbox' | 'surge') {
  const endpoint = outputType === 'surge' ? 'surge' : 'generate';
  const url = new URL(`/open/${endpoint}/${encodeURIComponent(record.code)}`, getOpenAPIOrigin());
  url.searchParams.set('token', record.token);
  return url.toString();
}

export default function DeviceManage() {
  const [devices, setDevices] = useState<Device[]>([]);
  const [inbounds, setInbounds] = useState<Inbound[]>([]);
  const [wireGuards, setWireGuards] = useState<WireGuard[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [deviceSaving, setDeviceSaving] = useState(false);
  const [bindingSaving, setBindingSaving] = useState(false);
  const [deletingDeviceCode, setDeletingDeviceCode] = useState<string | null>(null);
  const [bindingLoadingCode, setBindingLoadingCode] = useState<string | null>(null);

  const [deviceVisible, setDeviceVisible] = useState(false);
  const [deviceTitle, setDeviceTitle] = useState('新增设备');
  const [editingDevice, setEditingDevice] = useState<Partial<Device> | null>(null);
  const [deviceForm] = Form.useForm<Device>();

  const [bindingVisible, setBindingVisible] = useState(false);
  const [bindingDevice, setBindingDevice] = useState<Device | null>(null);
  const [selectedInboundTags, setSelectedInboundTags] = useState<string[]>([]);
  const [bindingSortByTag, setBindingSortByTag] = useState<Record<string, number>>({});

  // 设备页依赖设备、入站模板、WireGuard 模板三份基础数据，因此统一并行加载。
  const loadBaseData = useCallback(async (manual = false) => {
    try {
      if (manual) {
        setRefreshing(true);
      } else {
        setLoading(true);
      }
      const [deviceRes, inboundRes, wireGuardRes] = await Promise.all([
        api.getDevices(),
        api.getInbounds(),
        api.getWireGuards(),
      ]);
      setDevices(Array.isArray(deviceRes.data) ? deviceRes.data : []);
      setInbounds(Array.isArray(inboundRes.data) ? inboundRes.data : []);
      setWireGuards(Array.isArray(wireGuardRes.data) ? wireGuardRes.data : []);
    } catch {
      Message.error('加载设备管理数据失败');
      setDevices([]);
      setInbounds([]);
      setWireGuards([]);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => {
    loadBaseData();
  }, [loadBaseData]);

  useEffect(() => {
    if (!deviceVisible) {
      return;
    }
    deviceForm.resetFields();
    if (editingDevice) {
      deviceForm.setFieldsValue(editingDevice);
    }
  }, [deviceVisible, editingDevice, deviceForm]);

  // 绑定抽屉里按 sort 排序展示 Inbound，便于和最终输出顺序保持一致。
  const sortedInbounds = useMemo(() => [...inbounds].sort((left, right) => left.sort - right.sort), [inbounds]);

  // 每次关闭或切换设备时清空绑定编辑态，避免把上一台设备的选择残留到下一次。
  const resetBindingState = () => {
    setSelectedInboundTags([]);
    setBindingSortByTag({});
  };

  const handleAddDevice = () => {
    setEditingDevice({
      enabled: true,
      sort: devices.length,
      wireGuardTag: '',
      wireGuardClientAddr: '',
      wireGuardClientKey: '',
    });
    setDeviceTitle('新增设备');
    setDeviceVisible(true);
  };

  const handleEditDevice = (record: Device) => {
    setEditingDevice(record);
    setDeviceTitle('编辑设备');
    setDeviceVisible(true);
  };

  // 保存前统一 trim 字段，减少因为首尾空格导致的 tag / code / token 不一致问题。
  const handleSaveDevice = async () => {
    try {
      setDeviceSaving(true);
      const values = await deviceForm.validate();
      const payload: Device = {
        code: values.code.trim(),
        name: values.name.trim(),
        description: values.description?.trim() || '',
        token: values.token.trim(),
        enabled: !!values.enabled,
        sort: values.sort ?? 0,
        wireGuardTag: values.wireGuardTag?.trim() || '',
        wireGuardClientAddr: values.wireGuardClientAddr?.trim() || '',
        wireGuardClientKey: values.wireGuardClientKey?.trim() || '',
      };

      if (editingDevice?.code) {
        await api.updateDevice(editingDevice.code, payload);
      } else {
        await api.createDevice(payload);
      }

      Message.success('设备保存成功');
      setDeviceVisible(false);
      setEditingDevice(null);
      await loadBaseData();
    } catch (error) {
      if (error instanceof Error && error.message) {
        return;
      }
      Message.error('设备保存失败');
    } finally {
      setDeviceSaving(false);
    }
  };

  const handleDeleteDevice = async (record: Device) => {
    Modal.confirm({
      title: '确认删除设备',
      content: buildDeleteConfirmContent('设备', record.code, '设备的 Inbound 绑定也会一并删除。'),
      onOk: async () => {
        try {
          setDeletingDeviceCode(record.code);
          await api.deleteDevice(record.code);
          if (bindingDevice?.code === record.code) {
            setBindingVisible(false);
            setBindingDevice(null);
            resetBindingState();
          }
          if (editingDevice?.code === record.code) {
            setDeviceVisible(false);
            setEditingDevice(null);
          }
          Message.success('设备删除成功');
          await loadBaseData();
        } catch {
          Message.error('设备删除失败');
        } finally {
          setDeletingDeviceCode(null);
        }
      },
    });
  };

  const handleCopyConfigURL = async (record: Device, outputType: 'singbox' | 'surge') => {
    try {
      await navigator.clipboard.writeText(buildDeviceConfigURL(record, outputType));
      Message.success(`${outputType === 'surge' ? 'Surge' : 'sing-box'} 链接已复制`);
    } catch {
      Message.error('复制配置链接失败');
    }
  };

  // 打开绑定抽屉时，把后端列表转换成“勾选集合 + sort 映射”两份前端状态。
  const handleEditBindings = async (record: Device) => {
    try {
      setBindingLoadingCode(record.code);
      const response = await api.getDeviceInbounds(record.code);
      const bindings = Array.isArray(response.data) ? response.data : [];
      const state = buildDeviceInboundSelection(bindings);
      setBindingDevice(record);
      setSelectedInboundTags(state.selectedTags);
      setBindingSortByTag(state.sortByTag);
      setBindingVisible(true);
    } catch {
      Message.error('加载设备 Inbound 绑定失败');
      resetBindingState();
    } finally {
      setBindingLoadingCode(null);
    }
  };

  const handleToggleInbound = (tag: string, checked: boolean) => {
    setSelectedInboundTags((previous) => {
      if (checked) {
        return previous.includes(tag) ? previous : [...previous, tag];
      }
      return previous.filter((item) => item !== tag);
    });
  };

  const handleBindingSortChange = (tag: string, value?: number) => {
    setBindingSortByTag((previous) => ({
      ...previous,
      [tag]: value ?? 0,
    }));
  };

  // 绑定接口采用全量替换，所以这里始终重新生成完整 payload 提交。
  const handleSaveBindings = async () => {
    if (!bindingDevice) {
      return;
    }

    try {
      setBindingSaving(true);
      const payload = buildDeviceInboundPayload(bindingDevice.code, selectedInboundTags, bindingSortByTag);
      await api.setDeviceInbounds(bindingDevice.code, payload);
      Message.success('设备 Inbound 绑定已更新');
      setBindingVisible(false);
      setBindingDevice(null);
      resetBindingState();
    } catch {
      Message.error('保存设备 Inbound 绑定失败');
    } finally {
      setBindingSaving(false);
    }
  };

  return (
    <>
      <PageToolbar
        summary="设备管理：维护设备身份、WireGuard 绑定和设备可用入站。"
        count={devices.length}
        countLabel="台设备"
        onRefresh={() => loadBaseData(true)}
        refreshing={refreshing}
        onPrimaryAction={handleAddDevice}
        primaryActionLabel="新增设备"
      />

      <DataState
        loading={loading}
        isEmpty={devices.length === 0}
        emptyTitle="还没有设备"
        emptyDescription="先创建设备身份，再继续配置令牌、WireGuard 和入站绑定。"
        createLabel="立即新增设备"
        onCreate={handleAddDevice}
      >
        <Row gutter={[16, 16]}>
          {devices.map((record) => (
            <Col key={record.code} xs={24} sm={12} lg={6}>
              <div className="glass-card">
                <div className="card-header" style={{ padding: '12px 16px 8px' }}>
                  <div style={{ flex: 1 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                      <Title heading={6} style={{ margin: 0, fontSize: 15 }}>{record.name}</Title>
                      <Tag size="small" color={record.enabled ? 'arcoblue' : 'gray'} style={{ scale: '0.8', transformOrigin: 'left center' }}>
                        {record.enabled ? '启用' : '禁用'}
                      </Tag>
                    </div>
                    <Text type="secondary" style={{ marginTop: 2, display: 'block', fontSize: 11 }}>
                      {record.code}
                    </Text>
                  </div>
                  <div style={{ fontSize: 11, fontWeight: 700, opacity: 0.4 }}>
                    #{record.sort}
                  </div>
                </div>
                <div className="card-content" style={{ padding: '0 16px 12px' }}>
                  <Space direction="vertical" size={8} style={{ width: '100%' }}>
                    {record.description && (
                      <Text type="secondary" style={{ display: 'block', fontSize: 12, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {record.description}
                      </Text>
                    )}
                    <div style={{ background: 'rgba(0,0,0,0.02)', padding: '8px 12px', borderRadius: 10, border: '1px solid rgba(0,0,0,0.03)' }}>
                      <div style={{ fontSize: 10, color: 'var(--text-secondary)', marginBottom: 2, textTransform: 'uppercase' }}>WireGuard</div>
                      <div style={{ fontSize: 12, fontWeight: 500, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {getWireGuardLabel(record, wireGuards)}
                      </div>
                    </div>
                    <div style={{ display: 'flex', gap: 8 }}>
                      <Button size="small" style={{ flex: 1 }} onClick={() => handleCopyConfigURL(record, 'singbox')}>
                        sing-box
                      </Button>
                      <Button size="small" style={{ flex: 1 }} onClick={() => handleCopyConfigURL(record, 'surge')}>
                        Surge
                      </Button>
                    </div>
                  </Space>
                </div>
                <div className="card-footer" style={{ padding: '8px 16px' }}>
                  <div className="card-action-row">
                    <Button type="text" size="small" style={{ fontSize: 12, padding: '0 4px' }} onClick={() => handleEditDevice(record)}>编辑</Button>
                    <Button
                      type="text"
                      size="small"
                      loading={bindingLoadingCode === record.code}
                      style={{ fontSize: 12, padding: '0 4px' }}
                      onClick={() => handleEditBindings(record)}
                    >
                      入站
                    </Button>
                    <Button
                      type="text"
                      size="small"
                      status="danger"
                      loading={deletingDeviceCode === record.code}
                      style={{ fontSize: 12, padding: '0 4px' }}
                      onClick={() => handleDeleteDevice(record)}
                    >
                      删除
                    </Button>
                  </div>
                </div>
              </div>
            </Col>
          ))}
        </Row>
      </DataState>

      <Modal
        visible={deviceVisible}
        title={deviceTitle}
        style={{ width: 640 }}
        confirmLoading={deviceSaving}
        onOk={handleSaveDevice}
        onCancel={() => {
          setDeviceVisible(false);
          setEditingDevice(null);
        }}
        afterClose={() => deviceForm.resetFields()}
      >
        <Form form={deviceForm} layout="vertical">
          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <FormItem field="code" label="设备标识" rules={[{ required: true, message: '请输入设备标识' }]}>
                <Input placeholder="如 phone / tv / office" />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="name" label="设备名称" rules={[{ required: true, message: '请输入设备名称' }]}>
                <Input placeholder="如 手机 / 客厅电视" />
              </FormItem>
            </Col>
            <Col span={24}>
              <FormItem field="description" label="说明">
                <TextArea autoSize={{ minRows: 2, maxRows: 3 }} placeholder="记录设备用途等" />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="token" label="访问令牌" rules={[{ required: true, message: '请输入访问令牌' }]}>
                <Input placeholder="用于设备级生成校验" />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="sort" label="排序">
                <InputNumber style={{ width: '100%' }} />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="wireGuardTag" label="绑定 WireGuard Tag">
                <Input
                  placeholder="可留空"
                />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="enabled" label="是否启用" triggerPropName="checked">
                <Switch checkedText="启用" uncheckedText="停用" />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="wireGuardClientAddr" label="WG 客户端地址">
                <Input placeholder="如 172.19.0.2/32" />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="wireGuardClientKey" label="WG 客户端私钥">
                <Input placeholder="设备端私钥" />
              </FormItem>
            </Col>
          </Row>
        </Form>
      </Modal>

      <Drawer
        width={640}
        title={bindingDevice ? `${bindingDevice.name} 的 Inbound 绑定` : 'Inbound 绑定'}
        visible={bindingVisible}
        confirmLoading={bindingSaving}
        onOk={handleSaveBindings}
        onCancel={() => {
          setBindingVisible(false);
          setBindingDevice(null);
          resetBindingState();
        }}
        okText="保存绑定"
      >
        <div style={{ marginBottom: 12 }} className="inline-summary">
          勾选该设备需要启用的 Inbound 入站。
        </div>
        <Grid.Row gutter={[12, 12]}>
          {sortedInbounds.map((record) => (
            <Grid.Col key={record.tag} span={24}>
              <div style={{ 
                background: 'rgba(255,255,255,0.5)', 
                padding: '10px 16px', 
                borderRadius: 12, 
                border: '1px solid var(--border-color)',
                display: 'flex',
                alignItems: 'center',
                gap: 12
              }}>
                <Checkbox
                  checked={selectedInboundTags.includes(record.tag)}
                  onChange={(checked) => handleToggleInbound(record.tag, checked)}
                />
                <div style={{ flex: 1 }}>
                  <div style={{ fontSize: 13, fontWeight: 600 }}>{record.name}</div>
                  <div style={{ fontSize: 11, color: 'var(--text-secondary)' }}>{record.tag} ({record.type})</div>
                </div>
                <div style={{ width: 80 }}>
                  <InputNumber
                    size="mini"
                    style={{ width: '100%', borderRadius: 6 }}
                    value={bindingSortByTag[record.tag] ?? record.sort}
                    onChange={(value) => handleBindingSortChange(record.tag, value)}
                  />
                </div>
              </div>
            </Grid.Col>
          ))}
        </Grid.Row>
      </Drawer>
    </>
  );
}
