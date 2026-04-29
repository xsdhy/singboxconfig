import { Modal, Form, Input, Select, InputNumber, Message, Grid } from '@arco-design/web-react';
import { useEffect, useRef, useState } from 'react';
import Editor from '@monaco-editor/react';
import type { RuleSet, NodeGroup } from '../types';
import { prettyJsonText } from '../utils/json';

const FormItem = Form.Item;

interface Props {
  visible: boolean;
  title: string;
  initialValues: Partial<RuleSet>;
  nodeGroups: NodeGroup[];
  confirmLoading?: boolean;
  onOk: (values: RuleSet) => void;
  onCancel: () => void;
}

export default function RuleSetModal({ visible, title, initialValues, nodeGroups, confirmLoading, onOk, onCancel }: Props) {
  const [form] = Form.useForm();
  const [ruleSetType, setRuleSetType] = useState(initialValues.ruleSetType || 'local');
  const [editorValue, setEditorValue] = useState('{}');
  const editorRef = useRef<{ getValue: () => string } | null>(null);

  useEffect(() => {
    if (visible) {
      form.resetFields();
      form.setFieldsValue(initialValues);
      setRuleSetType(initialValues.ruleSetType || 'local');
      setEditorValue(prettyJsonText(initialValues.content));
    }
  }, [visible, initialValues, form]);

  const handleOk = async () => {
    const values = await form.validate();

    if (!values.name) {
      Message.error('请输入规则集名称');
      return;
    }
    if (!values.tag) {
      Message.error('请输入规则集Tag');
      return;
    }
    if (values.ruleSetType === 'remote' && !values.url) {
      Message.error('远程规则集必须填写URL');
      return;
    }

    if (values.ruleSetType === 'local') {
      try {
        const editorContent = editorRef.current?.getValue() || editorValue;
        JSON.parse(editorContent);
        values.content = editorContent;
      } catch {
        Message.error('JSON 格式不正确');
        return;
      }
    } else {
      values.content = '';
    }

    onOk(values);
  };

  const outboundOptions = [
    { label: 'direct', value: 'direct' },
    ...nodeGroups.map(g => ({ label: g.name, value: g.tag })),
  ];

  const downloadDetourOptions = [
    { label: 'direct', value: 'direct' },
    ...nodeGroups.map(g => ({ label: g.name, value: g.tag })),
  ];

  return (
    <Modal
      visible={visible}
      title={title}
      confirmLoading={confirmLoading}
      onOk={handleOk}
      onCancel={onCancel}
      style={{ width: 960 }}
      afterClose={() => form.resetFields()}
    >
      <Form form={form} layout="vertical" onValuesChange={(changed) => {
        if ('ruleSetType' in changed) {
          setRuleSetType(changed.ruleSetType as string);
        }
      }}>
        <Grid.Row gutter={24}>
          <Grid.Col xs={24} sm={12}>
            <FormItem field="name" label="规则集名称" rules={[{ required: true, message: '请输入规则集名称' }]}>
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
            <FormItem field="ruleSetType" label="类型" rules={[{ required: true, message: '请选择类型' }]}>
              <Select options={[
                { label: 'local', value: 'local' },
                { label: 'remote', value: 'remote' },
              ]} />
            </FormItem>
          </Grid.Col>
          <Grid.Col xs={24} sm={12}>
            <FormItem field="format" label="格式" rules={[{ required: true, message: '请选择格式' }]}>
              <Select options={[
                { label: 'source', value: 'source' },
                { label: 'binary', value: 'binary' },
              ]} />
            </FormItem>
          </Grid.Col>
        </Grid.Row>
        {ruleSetType === 'remote' && (
          <Grid.Row gutter={24}>
            <Grid.Col xs={24} sm={12}>
              <FormItem field="url" label="规则集URL" rules={[{ required: true, message: '请输入规则集 URL' }]}>
                <Input />
              </FormItem>
            </Grid.Col>
            <Grid.Col xs={24} sm={12}>
              <FormItem field="downloadDetour" label="下载">
                <Select options={downloadDetourOptions} />
              </FormItem>
            </Grid.Col>
          </Grid.Row>
        )}
        <Grid.Row gutter={24}>
          <Grid.Col xs={24} sm={12}>
            <FormItem field="outbound" label="出站">
              <Select options={outboundOptions} allowCreate showSearch />
            </FormItem>
          </Grid.Col>
          <Grid.Col xs={24} sm={12}>
            <FormItem field="sort" label="排序">
              <InputNumber style={{ width: '100%' }} />
            </FormItem>
          </Grid.Col>
        </Grid.Row>
        <FormItem field="ableDevices" label="支持设备">
          <Input placeholder="多个设备用逗号分隔" />
        </FormItem>
        {ruleSetType === 'local' && (
          <FormItem label="内容">
            <div className="editor-frame" style={{ height: 400 }}>
              <Editor
                height="100%"
                language="json"
                value={editorValue}
                onChange={(value) => setEditorValue(value || '')}
                onMount={(editor) => { editorRef.current = editor; }}
                options={{ minimap: { enabled: false }, fontSize: 14 }}
              />
            </div>
          </FormItem>
        )}
      </Form>
    </Modal>
  );
}
