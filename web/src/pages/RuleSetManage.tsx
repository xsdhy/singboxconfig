import { useState, useEffect, useCallback } from 'react';
import { Message } from '@arco-design/web-react';
import RuleSetTable from '../components/RuleSetTable';
import RuleSetModal from '../components/RuleSetModal';
import PageToolbar from '../components/PageToolbar';
import DataState from '../components/DataState';
import * as api from '../api';
import type { RuleSet, NodeGroup } from '../types';

export default function RuleSetManage() {
  const [ruleSets, setRuleSets] = useState<RuleSet[]>([]);
  const [nodeGroups, setNodeGroups] = useState<NodeGroup[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [deletingKey, setDeletingKey] = useState<string | null>(null);
  const [rsVisible, setRsVisible] = useState(false);
  const [rsTitle, setRsTitle] = useState('添加规则集');
  const [rsForm, setRsForm] = useState<Partial<RuleSet>>({});

  // 规则集列表依赖节点分组数据，因为弹窗里需要选择默认出站和下载出口。
  const loadData = useCallback(async (manual = false) => {
    try {
      if (manual) {
        setRefreshing(true);
      } else {
        setLoading(true);
      }
      const [rsRes, ngRes] = await Promise.all([
        api.getRuleSets(),
        api.getNodeGroups(),
      ]);
      setRuleSets(Array.isArray(rsRes.data) ? rsRes.data : []);
      setNodeGroups(Array.isArray(ngRes.data) ? ngRes.data : []);
    } catch {
      Message.error('加载规则集数据失败');
      setRuleSets([]);
      setNodeGroups([]);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => { loadData(); }, [loadData]);

  const handleAddRs = () => {
    setRsForm({ ruleSetType: 'local', format: 'source', url: '', downloadDetour: '', ableDevices: '', sort: 0, content: {} });
    setRsTitle('添加规则集');
    setRsVisible(true);
  };
  const handleEditRs = (r: RuleSet) => {
    let content = r.content;
    if (typeof content === 'string') {
      try { content = JSON.parse(content); } catch { /* keep as-is */ }
    }
    setRsForm({
      ...r,
      ruleSetType: r.ruleSetType || 'local',
      format: r.format || 'source',
      url: r.url || '',
      downloadDetour: r.downloadDetour || '',
      ableDevices: r.ableDevices || '',
      sort: r.sort || 0,
      content: content || {},
    });
    setRsTitle('编辑规则集');
    setRsVisible(true);
  };
  // handleRsOk 在提交前把 local 规则集的对象内容重新序列化为 JSON 字符串。
  const handleRsOk = async (values: RuleSet) => {
    try {
      setSubmitting(true);
      const formData = { ...rsForm, ...values };
      if (formData.ruleSetType === 'local' && formData.content && typeof formData.content === 'object') {
        formData.content = JSON.stringify(formData.content);
      }
      if (rsForm.tag) {
        await api.updateRuleSet(rsForm.tag, formData as RuleSet);
      } else {
        await api.createRuleSet(formData as RuleSet);
      }
      Message.success('保存成功');
      setRsVisible(false);
      await loadData();
    } catch { Message.error('保存失败'); }
    finally { setSubmitting(false); }
  };
  const handleDeleteRs = async (r: RuleSet) => {
    try {
      setDeletingKey(r.tag);
      await api.deleteRuleSet(r.tag);
      Message.success('删除成功');
      await loadData();
    } catch { Message.error('删除失败'); }
    finally { setDeletingKey(null); }
  };

  return (
    <>
      <PageToolbar
        summary="规则集：配置分流规则和远程规则集。"
        count={ruleSets.length}
        countLabel="个规则集"
        onRefresh={() => loadData(true)}
        refreshing={refreshing}
        onPrimaryAction={handleAddRs}
        primaryActionLabel="添加规则集"
      />

      <div style={{ paddingTop: 8 }}>
        <DataState
          loading={loading}
          isEmpty={ruleSets.length === 0}
          emptyTitle="还没有规则集"
          emptyDescription="可以先添加本地规则集，再根据需要补充远程 URL、格式和出站关联。"
          createLabel="立即添加规则集"
          onCreate={handleAddRs}
        >
          <RuleSetTable
            data={ruleSets}
            deletingKey={deletingKey}
            onEdit={handleEditRs}
            onDelete={handleDeleteRs}
          />
        </DataState>
      </div>

      <RuleSetModal visible={rsVisible} title={rsTitle} initialValues={rsForm} nodeGroups={nodeGroups}
        confirmLoading={submitting}
        onOk={handleRsOk} onCancel={() => setRsVisible(false)} />
    </>
  );
}
