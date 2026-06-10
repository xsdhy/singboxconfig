import { Modal, Form, Input, Select, Grid, Button, Space, Typography } from '@arco-design/web-react';
import { IconPlus, IconDelete } from '@arco-design/web-react/icon';
import { useEffect, useState } from 'react';
import type { NodeGroup, Device } from '../types';
import * as api from '../api';

const FormItem = Form.Item;
const TextArea = Input.TextArea;

// 分组类型可选项，前端与后端 entity.NodeGroupType 保持一致，避免裸字符串散落。
const GROUP_TYPE_OPTIONS = [
  { label: 'selector', value: 'selector' },
  { label: 'urltest', value: 'urltest' },
];

// deviceOverrideRow 表示一条设备级类型覆盖规则的编辑态。
interface deviceOverrideRow {
  deviceCode: string; // 设备编码。
  groupType: string; // 目标分组类型（selector / urltest）。
}

interface Props {
  visible: boolean;
  title: string;
  initialValues: Partial<NodeGroup>;
  confirmLoading?: boolean;
  onOk: (values: NodeGroup) => void;
  onCancel: () => void;
}

// parseOverrides 将后端的规则字符串解析为编辑态数组，解析规则与后端 ParseDeviceTypeOverrides 对齐：
// 逗号分隔多条规则，第一个冒号分隔设备编码与类型，忽略空白与不合法项。
function parseOverrides(raw?: string): deviceOverrideRow[] {
  if (!raw || !raw.trim()) return [];
  const rows: deviceOverrideRow[] = [];
  for (const rule of raw.split(',')) {
    const idx = rule.indexOf(':');
    if (idx < 0) continue;
    const deviceCode = rule.slice(0, idx).trim();
    const groupType = rule.slice(idx + 1).trim();
    if (!deviceCode || !groupType) continue;
    rows.push({ deviceCode, groupType });
  }
  return rows;
}

// serializeOverrides 将编辑态数组序列化回后端规则字符串，跳过设备编码为空的行。
function serializeOverrides(rows: deviceOverrideRow[]): string {
  return rows
    .filter((row) => row.deviceCode.trim() && row.groupType.trim())
    .map((row) => `${row.deviceCode.trim()}:${row.groupType.trim()}`)
    .join(',');
}

export default function NodeGroupModal({ visible, title, initialValues, confirmLoading, onOk, onCancel }: Props) {
  const [form] = Form.useForm();
  // overrides 与表单分开维护，关闭弹窗时一并复位。
  const [overrides, setOverrides] = useState<deviceOverrideRow[]>([]);
  const [devices, setDevices] = useState<Device[]>([]);

  useEffect(() => {
    if (visible) {
      form.resetFields();
      form.setFieldsValue(initialValues);
      setOverrides(parseOverrides(initialValues.deviceTypeOverrides));
      // 设备下拉数据复用现有设备列表接口，加载失败时静默降级为手动输入。
      api.getDevices()
        .then((res) => setDevices(Array.isArray(res.data) ? res.data : []))
        .catch(() => setDevices([]));
    }
  }, [visible, initialValues, form]);

  const handleOk = async () => {
    const values = await form.validate();
    // 把设备级覆盖编辑结果序列化回规则字符串后一并提交。
    onOk({ ...values, deviceTypeOverrides: serializeOverrides(overrides) });
  };

  const addOverrideRow = () => {
    setOverrides((prev) => [...prev, { deviceCode: '', groupType: 'urltest' }]);
  };

  const updateOverrideRow = (index: number, patch: Partial<deviceOverrideRow>) => {
    setOverrides((prev) => prev.map((row, i) => (i === index ? { ...row, ...patch } : row)));
  };

  const removeOverrideRow = (index: number) => {
    setOverrides((prev) => prev.filter((_, i) => i !== index));
  };

  return (
    <Modal
      visible={visible}
      title={title}
      confirmLoading={confirmLoading}
      onOk={handleOk}
      onCancel={onCancel}
      style={{ width: 640 }}
      afterClose={() => { form.resetFields(); setOverrides([]); }}
    >
      <Form form={form} layout="vertical">
        <Grid.Row gutter={24}>
          <Grid.Col xs={24} sm={12}>
            <FormItem field="name" label="分组名称" rules={[{ required: true, message: '请输入分组名称' }]}>
              <Input />
            </FormItem>
          </Grid.Col>
          <Grid.Col xs={24} sm={12}>
            <FormItem field="tag" label="Tag" rules={[{ required: true, message: '请输入 Tag' }]}>
              <Input />
            </FormItem>
          </Grid.Col>
        </Grid.Row>
        <Grid.Row gutter={24}>
          <Grid.Col xs={24} sm={12}>
            <FormItem field="groupType" label="分组类型" rules={[{ required: true, message: '请选择分组类型' }]}>
              <Select options={GROUP_TYPE_OPTIONS} />
            </FormItem>
          </Grid.Col>
          <Grid.Col xs={24} sm={12}>
            <FormItem field="testURL" label="测试URL">
              <Input />
            </FormItem>
          </Grid.Col>
        </Grid.Row>
        <FormItem field="include" label="包含节点">
          <TextArea autoSize={{ minRows: 3, maxRows: 6 }} />
        </FormItem>
        <FormItem field="exclude" label="排除节点">
          <TextArea autoSize={{ minRows: 3, maxRows: 6 }} />
        </FormItem>
        <FormItem
          label="设备级类型覆盖"
          // 未配置覆盖的设备会使用上方的默认分组类型，便于网关类设备走 urltest、终端设备走 selector。
          extra="为指定设备单独指定分组类型；未配置覆盖的设备使用上方默认分组类型。"
        >
          <Space direction="vertical" size={8} style={{ width: '100%' }}>
            {overrides.map((row, index) => (
              <Space key={index} size={8} style={{ width: '100%' }}>
                <Select
                  placeholder="选择设备"
                  showSearch
                  allowCreate
                  style={{ width: 240 }}
                  value={row.deviceCode || undefined}
                  onChange={(value) => updateOverrideRow(index, { deviceCode: value })}
                  options={devices.map((d) => ({ label: `${d.name || d.code}（${d.code}）`, value: d.code }))}
                />
                <Select
                  style={{ width: 140 }}
                  value={row.groupType}
                  onChange={(value) => updateOverrideRow(index, { groupType: value })}
                  options={GROUP_TYPE_OPTIONS}
                />
                <Button
                  type="text"
                  status="danger"
                  icon={<IconDelete />}
                  onClick={() => removeOverrideRow(index)}
                />
              </Space>
            ))}
            <Button type="dashed" long icon={<IconPlus />} onClick={addOverrideRow}>
              添加设备覆盖
            </Button>
            {overrides.length === 0 && (
              <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                暂无覆盖规则，所有设备都使用默认分组类型。
              </Typography.Text>
            )}
          </Space>
        </FormItem>
      </Form>
    </Modal>
  );
}
