import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  Button,
  Checkbox,
  Form,
  Grid,
  Input,
  InputNumber,
  Message,
  Modal,
  Pagination,
  Select,
  Space,
  Switch,
  Tag,
  Typography,
} from '@arco-design/web-react';
import Editor from '@monaco-editor/react';
import * as api from '../api';
import DataState from '../components/DataState';
import PageToolbar from '../components/PageToolbar';
import type { Outbound, Subscribe } from '../types';
import { buildDeleteConfirmContent } from '../utils/deleteConfirm';
import { normalizeJsonText, prettyJsonText } from '../utils/json';
import {
  buildBatchEnablePayload,
  buildOutboundQueryParams,
  createManualOutboundDraft,
  type OutboundFilters,
} from '../utils/outboundManagement';

const FormItem = Form.Item;
const TextArea = Input.TextArea;
const Search = Input.Search;
const Row = Grid.Row;
const Col = Grid.Col;
const { Title, Text } = Typography;
const PAGE_SIZE = 12;

function formatDateTime(value?: string | null) {
  if (!value) {
    return '未记录';
  }

  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return value;
  }

  return parsed.toLocaleString('zh-CN', { hour12: false });
}

export default function OutboundManage() {
  const [items, setItems] = useState<Outbound[]>([]);
  const [subscribes, setSubscribes] = useState<Subscribe[]>([]);
  const [filters, setFilters] = useState<OutboundFilters>({
    source: 'ALL',
    enabled: 'ALL',
    subscribeName: '',
    search: '',
  });
  const [searchInput, setSearchInput] = useState('');
  const [page, setPage] = useState(1);
  const [limit] = useState(PAGE_SIZE);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [batchLoading, setBatchLoading] = useState(false);
  const [deletingID, setDeletingID] = useState<number | null>(null);
  const [visible, setVisible] = useState(false);
  const [title, setTitle] = useState('新增 Outbound');
  const [editingItem, setEditingItem] = useState<Outbound | null>(null);
  const [editorValue, setEditorValue] = useState('{}');
  const [editorDirty, setEditorDirty] = useState(false);
  const [selectedIDs, setSelectedIDs] = useState<number[]>([]);
  const [form] = Form.useForm<Outbound>();
  const editorRef = useRef<{ getValue: () => string } | null>(null);

  // loadSubscribes 只负责加载筛选项，避免每次翻页都重复请求订阅列表。
  const loadSubscribes = useCallback(async () => {
    try {
      const response = await api.getSubscribes();
      setSubscribes(Array.isArray(response.data) ? response.data : []);
    } catch {
      Message.error('加载订阅筛选项失败');
      setSubscribes([]);
    }
  }, []);

  // loadData 统一处理筛选、分页和手动刷新，保证页面状态始终与服务端分页结果一致。
  const loadData = useCallback(async (options?: { manual?: boolean; nextPage?: number; nextFilters?: OutboundFilters }) => {
    const manual = options?.manual ?? false;
    const nextPage = options?.nextPage ?? page;
    const nextFilters = options?.nextFilters ?? filters;

    try {
      if (manual) {
        setRefreshing(true);
      } else {
        setLoading(true);
      }

      const response = await api.getOutbounds(buildOutboundQueryParams(nextFilters, nextPage, limit));
      const payload = response.data;
      setItems(Array.isArray(payload.items) ? payload.items : []);
      setTotal(payload.total ?? 0);
      setPage(payload.page ?? nextPage);
      setSelectedIDs([]);
    } catch {
      Message.error('加载 Outbound 列表失败');
      setItems([]);
      setTotal(0);
      setSelectedIDs([]);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [filters, limit, page]);

  useEffect(() => {
    void loadSubscribes();
  }, [loadSubscribes]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  useEffect(() => {
    if (!visible) {
      return;
    }

    form.resetFields();
    if (editingItem) {
      form.setFieldsValue(editingItem);
      setEditorValue(prettyJsonText(editingItem.configJson));
    } else {
      setEditorValue('{}');
    }
    setEditorDirty(false);
  }, [editingItem, form, visible]);

  const selectedCount = selectedIDs.length;
  const pageCount = useMemo(() => items.length, [items]);

  const applyFilters = async (nextFilters: OutboundFilters) => {
    setFilters(nextFilters);
    setPage(1);
    await loadData({ nextPage: 1, nextFilters });
  };

  const handleAdd = () => {
    setEditingItem(createManualOutboundDraft(total));
    setTitle('新增 Outbound');
    setVisible(true);
  };

  const handleEdit = (record: Outbound) => {
    if (record.source !== 'MANUAL') {
      Message.warning('订阅缓存节点只支持删除和批量启停，不支持直接编辑');
      return;
    }

    setEditingItem(record);
    setTitle('编辑 Outbound');
    setVisible(true);
  };

  const handleDelete = async (record: Outbound) => {
    if (!record.id) {
      Message.error('缺少 Outbound ID，无法删除');
      return;
    }

    Modal.confirm({
      title: '确认删除 Outbound',
      content: buildDeleteConfirmContent('Outbound', record.tag),
      onOk: async () => {
        try {
          setDeletingID(record.id!);
          await api.deleteOutbound(record.id!);
          Message.success('Outbound 删除成功');
          await loadData({ nextPage: page, nextFilters: filters });
        } catch {
          Message.error('Outbound 删除失败');
        } finally {
          setDeletingID(null);
        }
      },
    });
  };

  const handleSave = async () => {
    try {
      setSaving(true);
      const values = await form.validate();
      const payload: Outbound = {
        ...(editingItem ?? createManualOutboundDraft(total)),
        ...values,
        id: editingItem?.id,
        tag: values.tag.trim(),
        name: values.name.trim(),
        description: values.description?.trim() || '',
        type: values.type.trim(),
        enabled: !!values.enabled,
        sort: values.sort ?? 0,
        visibleDevices: values.visibleDevices?.trim() || '',
        configJson: normalizeJsonText(editorRef.current?.getValue() || editorValue),
        source: 'MANUAL',
        subscribeName: '',
      };

      if (editingItem?.id) {
        await api.updateOutbound(editingItem.id, payload);
      } else {
        await api.createOutbound(payload);
      }

      Message.success('Outbound 保存成功');
      setVisible(false);
      setEditingItem(null);
      await loadData({ nextPage: page, nextFilters: filters });
    } catch (error) {
      if (error instanceof Error && error.message) {
        return;
      }
      Message.error('Outbound 保存失败，请确认表单和 JSON 内容合法');
    } finally {
      setSaving(false);
    }
  };

  const handleBatchEnable = async (enabled: boolean) => {
    const payload = buildBatchEnablePayload(selectedIDs, enabled);
    if (payload.ids.length === 0) {
      Message.warning('请先选择至少一条 Outbound');
      return;
    }

    try {
      setBatchLoading(true);
      await api.batchEnableOutbounds(payload.ids, enabled);
      Message.success(enabled ? '批量启用成功' : '批量禁用成功');
      await loadData({ nextPage: page, nextFilters: filters });
    } catch {
      Message.error(enabled ? '批量启用失败' : '批量禁用失败');
    } finally {
      setBatchLoading(false);
    }
  };

  const toggleSelect = (recordID?: number, checked?: boolean) => {
    if (!recordID) {
      return;
    }

    setSelectedIDs((previous) => {
      const exists = previous.includes(recordID);
      if (checked && !exists) {
        return [...previous, recordID];
      }
      if (!checked && exists) {
        return previous.filter((id) => id !== recordID);
      }
      return previous;
    });
  };

  const allChecked = items.length > 0 && items.every((item) => item.id && selectedIDs.includes(item.id));
  const indeterminate = selectedIDs.length > 0 && !allChecked;

  return (
    <>
      <PageToolbar
        summary=""
        onRefresh={() => loadData({ manual: true, nextPage: page, nextFilters: filters })}
        refreshing={refreshing}
        onPrimaryAction={handleAdd}
        primaryActionLabel="新增 Outbound"
        extraContent={(
          <Space wrap>
            <Select
              value={filters.source}
              style={{ width: 130 }}
              onChange={(value) => void applyFilters({ ...filters, source: value as OutboundFilters['source'] })}
              options={[
                { label: '全部来源', value: 'ALL' },
                { label: '手工维护', value: 'MANUAL' },
                { label: '订阅缓存', value: 'SUBSCRIPTION' },
              ]}
            />
            <Select
              value={filters.enabled}
              style={{ width: 120 }}
              onChange={(value) => void applyFilters({ ...filters, enabled: value as OutboundFilters['enabled'] })}
              options={[
                { label: '全部状态', value: 'ALL' },
                { label: '仅启用', value: 'true' },
                { label: '仅禁用', value: 'false' },
              ]}
            />
            <Select
              allowClear
              placeholder="全部订阅"
              value={filters.subscribeName || undefined}
              style={{ width: 180 }}
              onChange={(value) => void applyFilters({ ...filters, subscribeName: String(value || '') })}
              options={subscribes.map((item) => ({ label: item.name, value: item.name }))}
            />
            <Search
              allowClear
              placeholder="按 Tag / 名称搜索"
              value={searchInput}
              style={{ width: 220 }}
              onChange={setSearchInput}
              onSearch={() => void applyFilters({ ...filters, search: searchInput })}
            />
            <Checkbox
              indeterminate={indeterminate}
              checked={allChecked}
              onChange={(checked) => {
                if (checked) {
                  setSelectedIDs(items.map((item) => item.id).filter((id): id is number => typeof id === 'number'));
                } else {
                  setSelectedIDs([]);
                }
              }}
            >
              已选 {selectedCount}
            </Checkbox>
            <Button size="small" loading={batchLoading} disabled={selectedCount === 0} onClick={() => void handleBatchEnable(true)}>
              批量启用
            </Button>
            <Button size="small" loading={batchLoading} disabled={selectedCount === 0} onClick={() => void handleBatchEnable(false)}>
              批量禁用
            </Button>
          </Space>
        )}
      />

      <DataState
        loading={loading}
        isEmpty={total === 0}
        emptyTitle="还没有 Outbound"
        emptyDescription="可以先新增一个手工 Outbound，或者切换筛选条件查看订阅缓存节点。"
        createLabel="立即新增 Outbound"
        onCreate={handleAdd}
      >
        <Row gutter={[16, 16]}>
          {items.map((record) => (
            <Col key={record.id ?? record.tag} xs={24} sm={12} lg={8} xl={6}>
              <div className="glass-card">
                <div className="card-header" style={{ padding: '12px 16px 8px', gap: 12 }}>
                  <Checkbox
                    checked={!!record.id && selectedIDs.includes(record.id)}
                    onChange={(checked) => toggleSelect(record.id, checked)}
                  />
                  <div style={{ flex: 1 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 6, flexWrap: 'wrap' }}>
                      <Title heading={6} style={{ margin: 0, fontSize: 15 }}>{record.name || record.tag}</Title>
                      <Tag size="small" color={record.enabled ? 'arcoblue' : 'gray'}>
                        {record.enabled ? '启用' : '禁用'}
                      </Tag>
                      <Tag size="small" color={record.source === 'MANUAL' ? 'orangered' : 'green'}>
                        {record.source === 'MANUAL' ? '手工' : '订阅'}
                      </Tag>
                    </div>
                    <div style={{ marginTop: 4, display: 'flex', gap: 6, flexWrap: 'wrap' }}>
                      <Tag size="small">{record.type || '未设置类型'}</Tag>
                      {record.subscribeName && <Tag size="small" color="blue">{record.subscribeName}</Tag>}
                    </div>
                  </div>
                  <div style={{ fontSize: 11, fontWeight: 700, opacity: 0.45 }}>
                    #{record.sort}
                  </div>
                </div>
                <div className="card-content" style={{ padding: '0 16px 12px' }}>
                  <Space direction="vertical" size={8} style={{ width: '100%' }}>
                    <div>
                      <Text type="secondary" style={{ fontSize: 11 }}>Tag</Text>
                      <div style={{ fontSize: 12, fontWeight: 600, wordBreak: 'break-all' }}>{record.tag}</div>
                    </div>
                    <div>
                      <Text type="secondary" style={{ fontSize: 11 }}>可见设备</Text>
                      <div style={{ fontSize: 12 }}>{record.visibleDevices || '全部设备'}</div>
                    </div>
                    <div>
                      <Text type="secondary" style={{ fontSize: 11 }}>说明</Text>
                      <div style={{ fontSize: 12, minHeight: 18 }}>{record.description || '无'}</div>
                    </div>
                    <div>
                      <Text type="secondary" style={{ fontSize: 11 }}>
                        {record.source === 'SUBSCRIPTION' ? '最近同步' : '更新时间'}
                      </Text>
                      <div style={{ fontSize: 12 }}>
                        {record.source === 'SUBSCRIPTION' ? formatDateTime(record.lastFetchTime) : formatDateTime(record.updatedAt)}
                      </div>
                    </div>
                  </Space>
                </div>
                <div className="card-footer" style={{ padding: '8px 16px' }}>
                  <div className="card-action-row">
                    {record.source === 'MANUAL' && (
                      <Button type="text" size="small" style={{ fontSize: 12, padding: '0 4px' }} onClick={() => handleEdit(record)}>
                        编辑
                      </Button>
                    )}
                    <Button
                      type="text"
                      size="small"
                      status="danger"
                      loading={deletingID === record.id}
                      style={{ fontSize: 12, padding: '0 4px' }}
                      onClick={() => void handleDelete(record)}
                    >
                      删除
                    </Button>
                  </div>
                </div>
              </div>
            </Col>
          ))}
        </Row>

        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 20, gap: 12, flexWrap: 'wrap' }}>
          <Text type="secondary" style={{ fontSize: 12 }}>
            当前页 {pageCount} 条，累计 {total} 条
          </Text>
          <Pagination
            sizeCanChange={false}
            current={page}
            pageSize={limit}
            total={total}
            onChange={(nextPage) => void loadData({ nextPage, nextFilters: filters })}
          />
        </div>
      </DataState>

      <Modal
        visible={visible}
        title={title}
        style={{ width: 860 }}
        confirmLoading={saving}
        onOk={() => void handleSave()}
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
                <Input placeholder="如 home-vpn" />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
                <Input placeholder="如 家宽直连" />
              </FormItem>
            </Col>
            <Col xs={24} sm={12}>
              <FormItem field="type" label="类型" rules={[{ required: true, message: '请输入类型' }]}>
                <Input placeholder="如 socks / vmess / hysteria2" />
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
            <Col xs={24} sm={12}>
              <FormItem label="来源">
                <Input value="MANUAL" disabled />
              </FormItem>
            </Col>
            <Col span={24}>
              <FormItem field="visibleDevices" label="可见设备">
                <Input placeholder="多个设备 code 用逗号分隔；留空表示全部设备可见" />
              </FormItem>
            </Col>
            <Col span={24}>
              <FormItem field="description" label="说明">
                <TextArea autoSize={{ minRows: 2, maxRows: 3 }} />
              </FormItem>
            </Col>
          </Row>
        </Form>

        <div style={{ marginTop: 12 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8, gap: 12, flexWrap: 'wrap' }}>
            <div>
              <Title heading={6} style={{ margin: 0 }}>Outbound JSON</Title>
              <Text type="secondary" style={{ fontSize: 12 }}>保存时会做 JSON 规范化，避免后端存入非法配置。</Text>
            </div>
            {editorDirty && (
              <div className="editor-status">
                <span className="editor-status-dot" />
                已修改
              </div>
            )}
          </div>
          <div className="editor-frame">
            <Editor
              height="340px"
              defaultLanguage="json"
              value={editorValue}
              onMount={(editor) => {
                editorRef.current = editor;
              }}
              onChange={(value) => {
                setEditorValue(value || '{}');
                setEditorDirty(true);
              }}
              options={{
                minimap: { enabled: false },
                fontSize: 13,
                lineNumbers: 'on',
                scrollBeyondLastLine: false,
              }}
            />
          </div>
        </div>
      </Modal>
    </>
  );
}
