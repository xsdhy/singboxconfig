import { useCallback, useEffect, useRef, useState } from 'react';
import { Button, Form, Grid, Input, InputNumber, Message, Modal, Space, Switch, Tag, Typography } from '@arco-design/web-react';
import Editor from '@monaco-editor/react';
import * as api from '../api';
import DataState from '../components/DataState';
import PageToolbar from '../components/PageToolbar';
import type { Inbound } from '../types';
import { buildDeleteConfirmContent } from '../utils/deleteConfirm';
import { normalizeJsonText, prettyJsonText } from '../utils/json';

const FormItem = Form.Item;
const TextArea = Input.TextArea;
const Row = Grid.Row;
const Col = Grid.Col;
const { Title, Text } = Typography;

export default function InboundManage() {
  const [items, setItems] = useState<Inbound[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [deletingKey, setDeletingKey] = useState<string | null>(null);
  const [visible, setVisible] = useState(false);
  const [title, setTitle] = useState('新增 Inbound');
  const [editingItem, setEditingItem] = useState<Partial<Inbound> | null>(null);
  const [editorDirty, setEditorDirty] = useState(false);
  const [editorValue, setEditorValue] = useState('{}');
  const [form] = Form.useForm<Inbound>();
  const editorRef = useRef<{ getValue: () => string } | null>(null);

  const loadData = useCallback(async (manual = false) => {
    try {
      if (manual) {
        setRefreshing(true);
      } else {
        setLoading(true);
      }
      const response = await api.getInbounds();
      setItems(Array.isArray(response.data) ? response.data : []);
    } catch {
      Message.error('加载 Inbound 列表失败');
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
    if (visible) {
      setEditorValue(prettyJsonText(editingItem?.configJson));
      setEditorDirty(false);
    }
  }, [visible, editingItem]);

  useEffect(() => {
    if (!visible) {
      return;
    }
    form.resetFields();
    if (editingItem) {
      form.setFieldsValue(editingItem);
    }
  }, [visible, editingItem, form]);

  const handleAdd = () => {
    setEditingItem({ enabled: true, sort: items.length, configJson: '{}' });
    setTitle('新增 Inbound');
    setVisible(true);
  };

  const handleEdit = (record: Inbound) => {
    setEditingItem(record);
    setTitle('编辑 Inbound');
    setVisible(true);
  };

  const handleDelete = async (record: Inbound) => {
    Modal.confirm({
      title: '确认删除 Inbound',
      content: buildDeleteConfirmContent('Inbound', record.tag, '相关设备上的 Inbound 绑定会同步清除。'),
      onOk: async () => {
        try {
          setDeletingKey(record.tag);
          await api.deleteInbound(record.tag);
          if (editingItem?.tag === record.tag) {
            setVisible(false);
            setEditingItem(null);
          }
          Message.success('Inbound 删除成功');
          await loadData();
        } catch {
          Message.error('Inbound 删除失败');
        } finally {
          setDeletingKey(null);
        }
      },
    });
  };

  const handleSave = async () => {
    try {
      setSaving(true);
      const values = await form.validate();
      const payload: Inbound = {
        tag: values.tag.trim(),
        name: values.name.trim(),
        description: values.description?.trim() || '',
        type: values.type.trim(),
        enabled: !!values.enabled,
        sort: values.sort ?? 0,
        configJson: normalizeJsonText(editorRef.current?.getValue() || editorValue),
      };

      if (editingItem?.tag) {
        await api.updateInbound(editingItem.tag, payload);
      } else {
        await api.createInbound(payload);
      }

      Message.success('Inbound 保存成功');
      setVisible(false);
      setEditingItem(null);
      await loadData();
    } catch (error) {
      if (error instanceof Error && error.message) {
        return;
      }
      Message.error('Inbound 保存失败，请确认 JSON 合法');
    } finally {
      setSaving(false);
    }
  };

  return (
    <>
      <PageToolbar
        summary="Inbound 模板：维护设备绑定的入站模板与原始 JSON 配置。"
        count={items.length}
        countLabel="个入站"
        onRefresh={() => loadData(true)}
        refreshing={refreshing}
        onPrimaryAction={handleAdd}
        primaryActionLabel="新增 Inbound"
      />

      <DataState
        loading={loading}
        isEmpty={items.length === 0}
        emptyTitle="还没有 Inbound"
        emptyDescription="先建立一个 Inbound 模板，设备页才能继续绑定入站能力。"
        createLabel="立即新增 Inbound"
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
                    <Tag size="small" style={{ marginTop: 4, scale: '0.85', transformOrigin: 'left center' }}>{record.type}</Tag>
                  </div>
                  <div style={{ fontSize: 11, fontWeight: 700, opacity: 0.4 }}>
                    #{record.sort}
                  </div>
                </div>
                <div className="card-content" style={{ padding: '0 16px 12px' }}>
                  <Space direction="vertical" size={4} style={{ width: '100%' }}>
                    <div>
                      <Text type="secondary" style={{ fontSize: 11 }}>Tag</Text>
                      <div style={{ fontSize: 12, fontWeight: 600 }}>{record.tag}</div>
                    </div>
                    {record.description && (
                      <Text type="secondary" style={{ display: 'block', fontSize: 11, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {record.description}
                      </Text>
                    )}
                  </Space>
                </div>
                <div className="card-footer" style={{ padding: '8px 16px' }}>
                  <div className="card-action-row">
                    <Button type="text" size="small" style={{ fontSize: 12, padding: '0 4px' }} onClick={() => handleEdit(record)}>编辑</Button>
                    <Button
                      type="text"
                      size="small"
                      status="danger"
                      loading={deletingKey === record.tag}
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
        visible={visible}
        title={title}
        style={{ width: 860 }}
        confirmLoading={saving}
        onOk={handleSave}
        onCancel={() => {
          setVisible(false);
          setEditingItem(null);
        }}
        afterClose={() => form.resetFields()}
      >
        <Form form={form} layout="vertical">
          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <FormItem field="tag" label="Tag" rules={[{ required: true, message: '请输入 Tag' }]}>
                <Input placeholder="如 tun / mixed / socks" />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
                <Input placeholder="如 TUN 入站" />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="type" label="类型" rules={[{ required: true, message: '请输入类型' }]}>
                <Input placeholder="如 tun / http / socks" />
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
            <Col span={24}>
              <FormItem label="JSON 配置">
                {editorDirty && (
                  <div style={{ marginBottom: 10 }}>
                    <div className="editor-status"><span className="editor-status-dot" />JSON 尚未保存</div>
                  </div>
                )}
                <div className="editor-frame" style={{ height: 300 }}>
                  <Editor
                    height="100%"
                    language="json"
                    value={editorValue}
                    onChange={(value) => {
                      setEditorValue(value || '');
                      setEditorDirty(true);
                    }}
                    onMount={(editor) => {
                      editorRef.current = editor;
                    }}
                    options={{ minimap: { enabled: false }, fontSize: 13, scrollBeyondLastLine: false }}
                  />
                </div>
              </FormItem>
            </Col>
          </Row>
        </Form>
      </Modal>
    </>
  );
}
