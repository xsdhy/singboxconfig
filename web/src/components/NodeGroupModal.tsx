import { Modal, Form, Input, Select, Grid } from '@arco-design/web-react';
import { useEffect } from 'react';
import type { NodeGroup } from '../types';

const FormItem = Form.Item;
const TextArea = Input.TextArea;

interface Props {
  visible: boolean;
  title: string;
  initialValues: Partial<NodeGroup>;
  confirmLoading?: boolean;
  onOk: (values: NodeGroup) => void;
  onCancel: () => void;
}

export default function NodeGroupModal({ visible, title, initialValues, confirmLoading, onOk, onCancel }: Props) {
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
              <Select options={[
                { label: 'selector', value: 'selector' },
                { label: 'urltest', value: 'urltest' },
              ]} />
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
      </Form>
    </Modal>
  );
}
