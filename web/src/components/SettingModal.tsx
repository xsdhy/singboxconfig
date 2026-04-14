import { Modal, Form, Input } from '@arco-design/web-react';
import { useEffect } from 'react';
import type { Setting } from '../types';

const FormItem = Form.Item;

interface Props {
  visible: boolean;
  title: string;
  initialValues: Partial<Setting>;
  confirmLoading?: boolean;
  onOk: (values: Setting) => void;
  onCancel: () => void;
}

export default function SettingModal({ visible, title, initialValues, confirmLoading, onOk, onCancel }: Props) {
  const [form] = Form.useForm();

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
        <FormItem field="key" label="键" rules={[{ required: true, message: '请输入设置键' }]}>
          <Input />
        </FormItem>
        <FormItem field="value" label="值" rules={[{ required: true, message: '请输入设置值' }]}>
          <Input />
        </FormItem>
      </Form>
    </Modal>
  );
}
