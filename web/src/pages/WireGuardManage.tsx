import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Button,
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
import type { WireGuard, WireGuardPeer } from '../types';
import { buildDeleteConfirmContent } from '../utils/deleteConfirm';

const FormItem = Form.Item;
const TextArea = Input.TextArea;
const Row = Grid.Row;
const Col = Grid.Col;
const { Title, Text } = Typography;

export default function WireGuardManage() {
  const [items, setItems] = useState<WireGuard[]>([]);
  const [peers, setPeers] = useState<WireGuardPeer[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [peerLoading, setPeerLoading] = useState(false);
  const [wireGuardSaving, setWireGuardSaving] = useState(false);
  const [peerSaving, setPeerSaving] = useState(false);
  const [deletingTag, setDeletingTag] = useState<string | null>(null);
  const [peerDeletingId, setPeerDeletingId] = useState<number | null>(null);
  const [peerDrawerOpeningTag, setPeerDrawerOpeningTag] = useState<string | null>(null);

  const [wireGuardVisible, setWireGuardVisible] = useState(false);
  const [wireGuardTitle, setWireGuardTitle] = useState('新增 WireGuard');
  const [editingItem, setEditingItem] = useState<Partial<WireGuard> | null>(null);
  const [wireGuardForm] = Form.useForm<WireGuard>();

  const [peerDrawerVisible, setPeerDrawerVisible] = useState(false);
  const [currentWireGuard, setCurrentWireGuard] = useState<WireGuard | null>(null);
  const [peerVisible, setPeerVisible] = useState(false);
  const [peerTitle, setPeerTitle] = useState('新增 Peer');
  const [editingPeer, setEditingPeer] = useState<Partial<WireGuardPeer> | null>(null);
  const [peerForm] = Form.useForm<WireGuardPeer>();

  const loadData = useCallback(async (manual = false) => {
    try {
      if (manual) {
        setRefreshing(true);
      } else {
        setLoading(true);
      }
      const response = await api.getWireGuards();
      setItems(Array.isArray(response.data) ? response.data : []);
    } catch {
      Message.error('加载 WireGuard 列表失败');
      setItems([]);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  useEffect(() => {
    if (!wireGuardVisible) {
      return;
    }
    wireGuardForm.resetFields();
    if (editingItem) {
      wireGuardForm.setFieldsValue(editingItem);
    }
  }, [wireGuardVisible, editingItem, wireGuardForm]);

  useEffect(() => {
    if (!peerVisible) {
      return;
    }
    peerForm.resetFields();
    if (editingPeer) {
      peerForm.setFieldsValue(editingPeer);
    }
  }, [peerVisible, editingPeer, peerForm]);

  const currentWireGuardTag = currentWireGuard?.tag ?? '';
  const sortedPeers = useMemo(() => [...peers].sort((left, right) => left.sort - right.sort), [peers]);

  const handleAdd = () => {
    setEditingItem({ enabled: true, sort: items.length, endpointTag: 'wg-ep', mtu: 1408 });
    setWireGuardTitle('新增 WireGuard');
    setWireGuardVisible(true);
  };

  const handleEdit = (record: WireGuard) => {
    setEditingItem(record);
    setWireGuardTitle('编辑 WireGuard');
    setWireGuardVisible(true);
  };

  const handleDelete = async (record: WireGuard) => {
    Modal.confirm({
      title: '确认删除 WireGuard',
      content: buildDeleteConfirmContent('WireGuard', record.tag, '相关 Peer 和设备上的 WireGuard 绑定会同步清除。'),
      onOk: async () => {
        try {
          setDeletingTag(record.tag);
          await api.deleteWireGuard(record.tag);
          if (currentWireGuard?.tag === record.tag) {
            setPeerDrawerVisible(false);
            setCurrentWireGuard(null);
            setPeers([]);
            setPeerVisible(false);
            setEditingPeer(null);
          }
          if (editingItem?.tag === record.tag) {
            setWireGuardVisible(false);
            setEditingItem(null);
          }
          Message.success('WireGuard 删除成功');
          await loadData();
        } catch {
          Message.error('WireGuard 删除失败');
        } finally {
          setDeletingTag(null);
        }
      },
    });
  };

  const handleSaveWireGuard = async () => {
    try {
      setWireGuardSaving(true);
      const values = await wireGuardForm.validate();
      const payload: WireGuard = {
        tag: values.tag.trim(),
        name: values.name.trim(),
        description: values.description?.trim() || '',
        enabled: !!values.enabled,
        sort: values.sort ?? 0,
        endpointTag: values.endpointTag.trim(),
        mtu: values.mtu ?? 0,
      };

      if (editingItem?.tag) {
        await api.updateWireGuard(editingItem.tag, payload);
      } else {
        await api.createWireGuard(payload);
      }

      Message.success('WireGuard 保存成功');
      setWireGuardVisible(false);
      setEditingItem(null);
      await loadData();
    } catch (error) {
      if (error instanceof Error && error.message) {
        return;
      }
      Message.error('WireGuard 保存失败');
    } finally {
      setWireGuardSaving(false);
    }
  };

  const handleOpenPeers = async (record: WireGuard) => {
    try {
      setPeerDrawerOpeningTag(record.tag);
      setPeerLoading(true);
      const response = await api.getWireGuardPeers(record.tag);
      setCurrentWireGuard(record);
      setPeers(Array.isArray(response.data) ? response.data : []);
      setPeerDrawerVisible(true);
    } catch {
      Message.error('加载 WireGuard Peer 失败');
      setPeers([]);
    } finally {
      setPeerDrawerOpeningTag(null);
      setPeerLoading(false);
    }
  };

  const reloadPeers = async (tag: string) => {
    const response = await api.getWireGuardPeers(tag);
    setPeers(Array.isArray(response.data) ? response.data : []);
  };

  const handleAddPeer = () => {
    setEditingPeer({
      wireGuardTag: currentWireGuardTag,
      enabled: true,
      sort: peers.length,
      port: 0,
      persistentKeepaliveInterval: 0,
      allowedIps: '',
      address: '',
      publicKey: '',
      preSharedKey: '',
    });
    setPeerTitle('新增 Peer');
    setPeerVisible(true);
  };

  const handleEditPeer = (record: WireGuardPeer) => {
    setEditingPeer(record);
    setPeerTitle('编辑 Peer');
    setPeerVisible(true);
  };

  const handleSavePeer = async () => {
    if (!currentWireGuardTag) {
      return;
    }

    try {
      setPeerSaving(true);
      const values = await peerForm.validate();
      const payload: WireGuardPeer = {
        id: values.id,
        wireGuardTag: currentWireGuardTag,
        sort: values.sort ?? 0,
        address: values.address.trim(),
        port: values.port ?? 0,
        publicKey: values.publicKey.trim(),
        preSharedKey: values.preSharedKey?.trim() || '',
        allowedIps: values.allowedIps.trim(),
        persistentKeepaliveInterval: values.persistentKeepaliveInterval ?? 0,
        enabled: !!values.enabled,
      };

      if (editingPeer?.id) {
        await api.updateWireGuardPeer(currentWireGuardTag, editingPeer.id, payload);
      } else {
        await api.createWireGuardPeer(currentWireGuardTag, payload);
      }

      Message.success('Peer 保存成功');
      setPeerVisible(false);
      setEditingPeer(null);
      await reloadPeers(currentWireGuardTag);
    } catch (error) {
      if (error instanceof Error && error.message) {
        return;
      }
      Message.error('Peer 保存失败');
    } finally {
      setPeerSaving(false);
    }
  };

  const handleDeletePeer = async (record: WireGuardPeer) => {
    if (!currentWireGuardTag || !record.id) {
      return;
    }
    const peerID = record.id;

    Modal.confirm({
      title: '确认删除 Peer',
      content: buildDeleteConfirmContent('Peer', `${record.address}:${record.port}`),
      onOk: async () => {
        try {
          setPeerDeletingId(peerID);
          await api.deleteWireGuardPeer(currentWireGuardTag, peerID);
          if (editingPeer?.id === peerID) {
            setPeerVisible(false);
            setEditingPeer(null);
          }
          Message.success('Peer 删除成功');
          await reloadPeers(currentWireGuardTag);
        } catch {
          Message.error('Peer 删除失败');
        } finally {
          setPeerDeletingId(null);
        }
      },
    });
  };

  return (
    <>
      <PageToolbar
        summary="WireGuard 模板与 Peer 分层维护，方便统一管理隧道模板和端点清单。"
        count={items.length}
        countLabel="个模板"
        onRefresh={() => loadData(true)}
        refreshing={refreshing}
        onPrimaryAction={handleAdd}
        primaryActionLabel="新增 WireGuard"
        extraContent={<span className="inline-summary" style={{ fontSize: 12 }}>模板与 Peer 分层维护</span>}
      />

      <DataState
        loading={loading}
        isEmpty={items.length === 0}
        emptyTitle="还没有 WireGuard 模板"
        emptyDescription="先建模板，再进入 Peer 抽屉维护地址、公钥和 AllowedIPs。"
        createLabel="立即新增 WireGuard"
        onCreate={handleAdd}
      >
        <Row gutter={[16, 16]}>
          {items.map((record) => (
            <Col key={record.tag} xs={24} sm={12} lg={6}>
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
                      {record.tag}
                    </Text>
                  </div>
                </div>
                <div className="card-content" style={{ padding: '0 16px 12px' }}>
                  <Space direction="vertical" size={6} style={{ width: '100%' }}>
                    <div style={{ background: 'rgba(0,0,0,0.02)', padding: '6px 10px', borderRadius: 8 }}>
                      <div style={{ fontSize: 10, color: 'var(--text-secondary)', textTransform: 'uppercase' }}>Endpoint Tag</div>
                      <div style={{ fontSize: 12, fontWeight: 500 }}>{record.endpointTag}</div>
                    </div>
                    <div style={{ display: 'flex', gap: 12 }}>
                      <div>
                        <Text type="secondary" style={{ fontSize: 10 }}>MTU</Text>
                        <div style={{ fontSize: 12, fontWeight: 500 }}>{record.mtu}</div>
                      </div>
                      <div>
                        <Text type="secondary" style={{ fontSize: 10 }}>排序</Text>
                        <div style={{ fontSize: 12, fontWeight: 500 }}>{record.sort}</div>
                      </div>
                    </div>
                  </Space>
                </div>
                <div className="card-footer" style={{ padding: '8px 16px' }}>
                  <div className="card-action-row">
                    <Button type="text" size="small" style={{ fontSize: 12, padding: '0 4px' }} onClick={() => handleEdit(record)}>编辑</Button>
                    <Button
                      type="text"
                      size="small"
                      loading={peerDrawerOpeningTag === record.tag}
                      style={{ fontSize: 12, padding: '0 4px' }}
                      onClick={() => handleOpenPeers(record)}
                    >
                      Peer
                    </Button>
                    <Button
                      type="text"
                      size="small"
                      status="danger"
                      loading={deletingTag === record.tag}
                      style={{ fontSize: 12, padding: '0 4px' }}
                      onClick={() => handleDelete(record)}
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
        visible={wireGuardVisible}
        title={wireGuardTitle}
        style={{ width: 640 }}
        confirmLoading={wireGuardSaving}
        onOk={handleSaveWireGuard}
        onCancel={() => {
          setWireGuardVisible(false);
          setEditingItem(null);
        }}
        afterClose={() => wireGuardForm.resetFields()}
      >
        <Form form={wireGuardForm} layout="vertical">
          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <FormItem field="tag" label="Tag" rules={[{ required: true, message: '请输入 Tag' }]}>
                <Input placeholder="如 wg-home" />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
                <Input placeholder="如 家庭 WireGuard" />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="endpointTag" label="Endpoint Tag" rules={[{ required: true, message: '请输入 Endpoint Tag' }]}>
                <Input placeholder="默认 wg-ep" />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="mtu" label="MTU">
                <InputNumber style={{ width: '100%' }} />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="sort" label="排序">
                <InputNumber style={{ width: '100%' }} />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="enabled" label="是否启用" triggerPropName="checked">
                <Switch checkedText="启用" uncheckedText="停用" />
              </FormItem>
            </Col>
            <Col span={24}>
              <FormItem field="description" label="说明">
                <TextArea autoSize={{ minRows: 2, maxRows: 3 }} />
              </FormItem>
            </Col>
          </Row>
        </Form>
      </Modal>

      <Drawer
        width={560}
        visible={peerDrawerVisible}
        title={currentWireGuard ? `${currentWireGuard.name} - Peer 明细` : 'Peer 明细'}
        onCancel={() => {
          setPeerDrawerVisible(false);
          setCurrentWireGuard(null);
          setPeers([]);
        }}
        footer={null}
      >
        <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
          <Button type="primary" size="small" onClick={handleAddPeer} style={{ borderRadius: 8 }}>
            新增 Peer
          </Button>
          <div className="inline-summary" style={{ fontSize: 12 }}>
            当前 {sortedPeers.length} 个 Peer
          </div>
        </div>
        
        {peerLoading && <div style={{ textAlign: 'center', padding: '20px 0', fontSize: 12 }}>加载中...</div>}
        
        <Row gutter={[12, 12]}>
          {sortedPeers.map((record) => (
            <Col key={record.id} span={24}>
              <div className="glass-card" style={{ height: 'auto' }}>
                <div className="card-header" style={{ padding: '12px 16px 8px' }}>
                  <div style={{ flex: 1 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                      <Title heading={6} style={{ margin: 0, fontSize: 14 }}>{record.address}:{record.port}</Title>
                      <Tag size="small" color={record.enabled ? 'arcoblue' : 'gray'} style={{ scale: '0.8', transformOrigin: 'left center' }}>
                        {record.enabled ? '启用' : '禁用'}
                      </Tag>
                    </div>
                    <Text type="secondary" style={{ marginTop: 2, display: 'block', fontSize: 10 }}>
                      ID: {record.id} | 排序: {record.sort}
                    </Text>
                  </div>
                </div>
                <div className="card-content" style={{ padding: '0 16px 12px' }}>
                  <Space direction="vertical" size={4} style={{ width: '100%' }}>
                    <div style={{ background: 'rgba(0,0,0,0.02)', padding: '8px 12px', borderRadius: 8 }}>
                      <div style={{ fontSize: 10, color: 'var(--text-secondary)', marginBottom: 2 }}>ALLOWED IPS</div>
                      <div style={{ fontSize: 12, fontWeight: 500, fontFamily: 'monospace', wordBreak: 'break-all' }}>{record.allowedIps}</div>
                    </div>
                    <div>
                      <Text type="secondary" style={{ fontSize: 10 }}>心跳间隔</Text>
                      <div style={{ fontSize: 12, fontWeight: 500 }}>{record.persistentKeepaliveInterval}s</div>
                    </div>
                  </Space>
                </div>
                <div className="card-footer" style={{ padding: '8px 16px' }}>
                  <div className="card-action-row">
                    <Button type="text" size="small" style={{ fontSize: 12, padding: '0 4px' }} onClick={() => handleEditPeer(record)}>编辑</Button>
                    <Button
                      type="text"
                      size="small"
                      status="danger"
                      loading={peerDeletingId === record.id}
                      style={{ fontSize: 12, padding: '0 4px' }}
                      onClick={() => handleDeletePeer(record)}
                    >
                      删除
                    </Button>
                  </div>
                </div>
              </div>
            </Col>
          ))}
        </Row>
      </Drawer>

      <Modal
        visible={peerVisible}
        title={peerTitle}
        style={{ width: 640 }}
        confirmLoading={peerSaving}
        onOk={handleSavePeer}
        onCancel={() => {
          setPeerVisible(false);
          setEditingPeer(null);
        }}
        afterClose={() => peerForm.resetFields()}
      >
        <Form form={peerForm} layout="vertical">
          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <FormItem field="address" label="地址" rules={[{ required: true, message: '请输入地址' }]}>
                <Input placeholder="Peer 服务端地址" />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="port" label="端口" rules={[{ required: true, message: '请输入端口' }]}>
                <InputNumber style={{ width: '100%' }} />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="publicKey" label="公钥" rules={[{ required: true, message: '请输入公钥' }]}>
                <Input placeholder="Peer 公钥" />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="preSharedKey" label="预共享密钥">
                <Input placeholder="可留空" />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="allowedIps" label="AllowedIPs" rules={[{ required: true, message: '请输入 AllowedIPs' }]}>
                <Input placeholder="多个网段用逗号分隔" />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="persistentKeepaliveInterval" label="PersistentKeepalive">
                <InputNumber style={{ width: '100%' }} />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="sort" label="排序">
                <InputNumber style={{ width: '100%' }} />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="enabled" label="是否启用" triggerPropName="checked">
                <Switch checkedText="启用" uncheckedText="停用" />
              </FormItem>
            </Col>
          </Row>
        </Form>
      </Modal>
    </>
  );
}
