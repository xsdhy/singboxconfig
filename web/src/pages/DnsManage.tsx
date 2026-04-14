import { useCallback, useEffect, useRef, useState } from 'react';
import { Button, Message, Modal, Space } from '@arco-design/web-react';
import Editor from '@monaco-editor/react';
import * as api from '../api';
import PageToolbar from '../components/PageToolbar';
import { DNS_SETTING_KEY, getDefaultDnsEditorText, getDnsEditorText } from '../utils/dns';
import { normalizeJsonText } from '../utils/json';


export default function DnsManage() {
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [dirty, setDirty] = useState(false);
  const [editorValue, setEditorValue] = useState(getDefaultDnsEditorText());
  const editorRef = useRef<{ getValue: () => string } | null>(null);

  const loadDnsConfig = useCallback(async () => {
    try {
      setLoading(true);
      const response = await api.getSettingByKey(DNS_SETTING_KEY);
      setEditorValue(getDnsEditorText(response.data.value));
      setDirty(false);
    } catch {
      setEditorValue(getDefaultDnsEditorText());
      setDirty(false);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadDnsConfig();
  }, [loadDnsConfig]);

  const handleSave = async () => {
    try {
      setSaving(true);
      const value = normalizeJsonText(editorRef.current?.getValue() || editorValue);
      try {
        await api.updateSetting(DNS_SETTING_KEY, { key: DNS_SETTING_KEY, value });
      } catch {
        await api.createSetting({ key: DNS_SETTING_KEY, value });
      }
      Message.success('DNS 配置保存成功');
      setEditorValue(getDnsEditorText(value));
      setDirty(false);
    } catch {
      Message.error('DNS 配置保存失败，请确认 JSON 合法');
    } finally {
      setSaving(false);
    }
  };

  const handleResetToDefault = () => {
    Modal.confirm({
      title: '恢复默认 DNS',
      content: '会使用内置默认 DNS 模板覆盖当前内容，未保存修改将丢失。',
      onOk: () => {
        setEditorValue(getDefaultDnsEditorText());
        setDirty(true);
      },
    });
  };

  return (
    <>
      <PageToolbar
        summary="读写 GlobalSettings[dns_config]，支持 Sing-Box 标准 DNS 格式。"
        extraContent={dirty ? <div className="editor-status"><span className="editor-status-dot" />未保存更改</div> : undefined}
        onRefresh={loadDnsConfig}
        refreshing={loading}
      />
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
        <div className="inline-summary" style={{ padding: '6px 12px', fontSize: 12 }}>
          保存前会校验 JSON；恢复默认只替换编辑区内容，不会立即提交到后端。
        </div>
        <Space size="small">
          <Button size="small" onClick={handleResetToDefault} style={{ borderRadius: 8 }}>恢复默认</Button>
          <Button size="small" type="primary" loading={saving} onClick={handleSave} style={{ borderRadius: 8 }}>保存 DNS 配置</Button>
        </Space>
      </div>
      <div className="editor-frame" style={{ height: 'calc(100vh - 220px)', minHeight: 400 }}>
        <Editor
          height="100%"
          language="json"
          value={editorValue}
          onChange={(value) => {
            setEditorValue(value || '');
            setDirty(true);
          }}
          onMount={(editor) => {
            editorRef.current = editor;
          }}
          options={{
            minimap: { enabled: false },
            fontSize: 13,
            scrollBeyondLastLine: false,
            automaticLayout: true,
          }}
        />
      </div>
    </>
  );
}
