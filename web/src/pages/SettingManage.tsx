import { useState, useEffect, useCallback } from 'react';
import { Message } from '@arco-design/web-react';
import SettingTable from '../components/SettingTable';
import SettingModal from '../components/SettingModal';
import PageToolbar from '../components/PageToolbar';
import DataState from '../components/DataState';
import * as api from '../api';
import type { Setting } from '../types';

export default function SettingManage() {
  const [settings, setSettings] = useState<Setting[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [deletingKey, setDeletingKey] = useState<string | null>(null);
  const [setVisible, setSetVisible] = useState(false);
  const [setTitle, setSetTitle] = useState('添加设置');
  const [setForm, setSetForm] = useState<Partial<Setting>>({});

  const loadData = useCallback(async (manual = false) => {
    try {
      if (manual) {
        setRefreshing(true);
      } else {
        setLoading(true);
      }
      const res = await api.getSettings();
      setSettings(Array.isArray(res.data) ? res.data : []);
    } catch {
      Message.error('加载设置数据失败');
      setSettings([]);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => { loadData(); }, [loadData]);

  const handleAddSet = () => { setSetForm({}); setSetTitle('添加设置'); setSetVisible(true); };
  const handleEditSet = (r: Setting) => { setSetForm({ ...r }); setSetTitle('编辑设置'); setSetVisible(true); };
  const handleSetOk = async (values: Setting) => {
    try {
      setSubmitting(true);
      if (setForm.key) {
        await api.updateSetting(setForm.key, { ...setForm, ...values } as Setting);
      } else {
        await api.createSetting(values);
      }
      Message.success('保存成功');
      setSetVisible(false);
      await loadData();
    } catch { Message.error('保存失败'); }
    finally { setSubmitting(false); }
  };
  const handleDeleteSet = async (r: Setting) => {
    try {
      setDeletingKey(r.key);
      await api.deleteSetting(r.key);
      Message.success('删除成功');
      await loadData();
    } catch { Message.error('删除失败'); }
    finally { setDeletingKey(null); }
  };

  return (
    <>
      <PageToolbar
        summary="全局设置：配置系统全局参数。"
        count={settings.length}
        countLabel="项设置"
        onRefresh={() => loadData(true)}
        refreshing={refreshing}
        onPrimaryAction={handleAddSet}
        primaryActionLabel="添加设置"
      />

      <div style={{ paddingTop: 8 }}>
        <DataState
          loading={loading}
          isEmpty={settings.length === 0}
          emptyTitle="还没有全局设置"
          emptyDescription="全局设置会影响多个功能模块，建议先录入系统级公共参数。"
          createLabel="立即添加设置"
          onCreate={handleAddSet}
        >
          <SettingTable
            data={settings}
            deletingKey={deletingKey}
            onEdit={handleEditSet}
            onDelete={handleDeleteSet}
          />
        </DataState>
      </div>

      <SettingModal visible={setVisible} title={setTitle} initialValues={setForm}
        confirmLoading={submitting}
        onOk={handleSetOk} onCancel={() => setSetVisible(false)} />
    </>
  );
}
