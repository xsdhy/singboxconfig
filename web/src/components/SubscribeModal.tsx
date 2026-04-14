import { Alert, Form, Grid, Input, InputNumber, Modal, Switch, Typography } from '@arco-design/web-react';
import { useEffect } from 'react';
import type { Subscribe } from '../types';

const FormItem = Form.Item;
const Row = Grid.Row;
const Col = Grid.Col;
const TextArea = Input.TextArea;
const { Text } = Typography;

interface Props {
  visible: boolean;
  title: string;
  initialValues: Partial<Subscribe>;
  confirmLoading?: boolean;
  onOk: (values: Subscribe) => void;
  onCancel: () => void;
}

export default function SubscribeModal({ visible, title, initialValues, confirmLoading, onOk, onCancel }: Props) {
  const [form] = Form.useForm<Subscribe>();

  useEffect(() => {
    if (visible) {
      form.resetFields();
      form.setFieldsValue(initialValues);
    }
  }, [visible, initialValues, form]);

  const handleOk = async () => {
    const values = await form.validate();
    onOk(values);
  };

  return (
    <Modal
      visible={visible}
      title={title}
      confirmLoading={confirmLoading}
      onOk={handleOk}
      onCancel={onCancel}
      style={{ width: 640 }}
      afterClose={() => form.resetFields()}
    >
      <Form form={form} layout="vertical">
        <Row gutter={16}>
          <Col xs={24} sm={12}>
            <FormItem field="name" label="订阅名称" rules={[{ required: true, message: '请输入订阅名称' }]}>
              <Input />
            </FormItem>
          </Col>
          <Col xs={24} sm={12}>
            <FormItem field="status" label="状态" triggerPropName="checked">
              <Switch checkedText="启用" uncheckedText="禁用" />
            </FormItem>
          </Col>
          <Col span={24}>
            <FormItem field="url" label="订阅地址" rules={[{ required: true, message: '请输入订阅地址' }]}>
              <Input />
            </FormItem>
          </Col>
          <Col xs={24} sm={12}>
            <FormItem field="userAgent" label="User-Agent">
              <Input placeholder="留空时由后端使用默认桌面浏览器 UA" />
            </FormItem>
          </Col>
          <Col xs={24} sm={12}>
            <FormItem field="visibleDevices" label="可见设备">
              <Input placeholder="多个设备 code 用逗号分隔；留空表示全部设备可见" />
            </FormItem>
          </Col>
        </Row>

        <div style={{ marginTop: 8, marginBottom: 16 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8, gap: 12, flexWrap: 'wrap' }}>
            <Text style={{ fontWeight: 600 }}>Outbound 缓存配置</Text>
            <Text type="secondary" style={{ fontSize: 12 }}>
              0 = 每次都实时拉取，不保留缓存
            </Text>
          </div>
          <Alert
            type="info"
            content="订阅基础信息和缓存配置走两条接口保存。这里统一在同一次提交里顺序执行，避免用户反复切换页面。"
            style={{ marginBottom: 12 }}
          />
          <FormItem field="outboundCacheDuration" label="缓存时长（分钟）">
            <InputNumber min={0} precision={0} style={{ width: '100%' }} />
          </FormItem>
          <TextArea
            value={[
              `最近拉取：${initialValues.outboundLastFetchTime || '未拉取'}`,
              `最近状态：${initialValues.outboundLastFetchStatus || '未记录'}`,
              initialValues.outboundLastFetchError ? `最近错误：${initialValues.outboundLastFetchError}` : '',
            ].filter(Boolean).join('\n')}
            autoSize={{ minRows: 3, maxRows: 4 }}
            disabled
          />
        </div>
      </Form>
    </Modal>
  );
}
