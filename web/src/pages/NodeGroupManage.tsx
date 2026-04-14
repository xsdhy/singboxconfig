import { useState, useEffect, useCallback } from 'react';
import { Message } from '@arco-design/web-react';
import NodeGroupTable from '../components/NodeGroupTable';
import NodeGroupModal from '../components/NodeGroupModal';
import PageToolbar from '../components/PageToolbar';
import DataState from '../components/DataState';
import * as api from '../api';
import type { NodeGroup } from '../types';

export default function NodeGroupManage() {
  const [nodeGroups, setNodeGroups] = useState<NodeGroup[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [deletingKey, setDeletingKey] = useState<string | null>(null);
  const [ngVisible, setNgVisible] = useState(false);
  const [ngTitle, setNgTitle] = useState('添加节点分组');
  const [ngForm, setNgForm] = useState<Partial<NodeGroup>>({});

  // loadData 统一处理首屏加载和工具栏手动刷新。
  const loadData = useCallback(async (manual = false) => {
    try {
      if (manual) {
        setRefreshing(true);
      } else {
        setLoading(true);
      }
      const res = await api.getNodeGroups();
      setNodeGroups(Array.isArray(res.data) ? res.data : []);
    } catch {
      Message.error('加载节点分组数据失败');
      setNodeGroups([]);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => { loadData(); }, [loadData]);

  const handleAddNg = () => { setNgForm({}); setNgTitle('添加节点分组'); setNgVisible(true); };
  const handleEditNg = (r: NodeGroup) => { setNgForm({ ...r }); setNgTitle('编辑节点分组'); setNgVisible(true); };
  // handleNgOk 复用同一套弹窗处理新增和编辑，仅根据是否已有 tag 决定调用哪条接口。
  const handleNgOk = async (values: NodeGroup) => {
    try {
      setSubmitting(true);
      if (ngForm.tag) {
        await api.updateNodeGroup(ngForm.tag, { ...ngForm, ...values } as NodeGroup);
      } else {
        await api.createNodeGroup(values);
      }
      Message.success('保存成功');
      setNgVisible(false);
      await loadData();
    } catch { Message.error('保存失败'); }
    finally { setSubmitting(false); }
  };
  const handleDeleteNg = async (r: NodeGroup) => {
    try {
      setDeletingKey(r.tag);
      await api.deleteNodeGroup(r.tag);
      Message.success('删除成功');
      await loadData();
    } catch { Message.error('删除失败'); }
    finally { setDeletingKey(null); }
  };

  return (
    <>
      <PageToolbar
        summary="节点分组：管理节点的分组和出站策略。"
        count={nodeGroups.length}
        countLabel="个分组"
        onRefresh={() => loadData(true)}
        refreshing={refreshing}
        onPrimaryAction={handleAddNg}
        primaryActionLabel="添加节点分组"
      />

      <div style={{ paddingTop: 8 }}>
        <DataState
          loading={loading}
          isEmpty={nodeGroups.length === 0}
          emptyTitle="还没有节点分组"
          emptyDescription="可以先建立urltest或selector分组，再逐步填入 include / exclude 条件。"
          createLabel="立即添加分组"
          onCreate={handleAddNg}
        >
          <NodeGroupTable
            data={nodeGroups}
            deletingKey={deletingKey}
            onEdit={handleEditNg}
            onDelete={handleDeleteNg}
          />
        </DataState>
      </div>

      <NodeGroupModal visible={ngVisible} title={ngTitle} initialValues={ngForm}
        confirmLoading={submitting}
        onOk={handleNgOk} onCancel={() => setNgVisible(false)} />
    </>
  );
}
