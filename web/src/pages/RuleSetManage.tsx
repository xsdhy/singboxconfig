import { useState, useEffect, useCallback } from 'react';
import { Message } from '@arco-design/web-react';
import RuleSetTable from '../components/RuleSetTable';
import RuleSetModal from '../components/RuleSetModal';
import RuleSetCopyURLModal from '../components/RuleSetCopyURLModal';
import PageToolbar from '../components/PageToolbar';
import DataState from '../components/DataState';
import * as api from '../api';
import { SYSTEM_HOST_SETTING_KEY } from '../utils/systemHost';
import type { RuleSet, NodeGroup, Device } from '../types';

export default function RuleSetManage() {
  const [ruleSets, setRuleSets] = useState<RuleSet[]>([]);
  const [nodeGroups, setNodeGroups] = useState<NodeGroup[]>([]);
  const [devices, setDevices] = useState<Device[]>([]);
  const [systemHost, setSystemHost] = useState('');
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [deletingKey, setDeletingKey] = useState<string | null>(null);
  const [togglingKey, setTogglingKey] = useState<string | null>(null);
  const [rsVisible, setRsVisible] = useState(false);
  const [rsTitle, setRsTitle] = useState('添加规则集');
  const [rsForm, setRsForm] = useState<Partial<RuleSet>>({});
  const [copyTarget, setCopyTarget] = useState<RuleSet | null>(null);

  // 规则集列表依赖节点分组数据，因为弹窗里需要选择默认出站和下载出口。
  // 同时加载设备与系统 Host，供「复制地址」弹窗拼接 open 接口的绝对地址。
  const loadData = useCallback(async (manual = false) => {
    try {
      if (manual) {
        setRefreshing(true);
      } else {
        setLoading(true);
      }
      const [rsRes, ngRes, deviceRes] = await Promise.all([
        api.getRuleSets(),
        api.getNodeGroups(),
        api.getDevices(),
      ]);
      setRuleSets(Array.isArray(rsRes.data) ? rsRes.data : []);
      setNodeGroups(Array.isArray(ngRes.data) ? ngRes.data : []);
      setDevices(Array.isArray(deviceRes.data) ? deviceRes.data : []);
    } catch {
      Message.error('加载规则集数据失败');
      setRuleSets([]);
      setNodeGroups([]);
      setDevices([]);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  // 系统 Host 单独读取：未配置时接口返回 404，此时视为空串即可，不影响列表加载。
  const loadSystemHost = useCallback(async () => {
    try {
      const res = await api.getSettingByKey(SYSTEM_HOST_SETTING_KEY);
      setSystemHost(res.data.value || '');
    } catch {
      setSystemHost('');
    }
  }, []);

  useEffect(() => { loadData(); loadSystemHost(); }, [loadData, loadSystemHost]);

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

  const handleChangeOutbound = async (record: RuleSet, newOutbound: string) => {
    try {
      setTogglingKey(record.tag);
      await api.updateRuleSet(record.tag, { ...record, outbound: newOutbound });
      Message.success('出站已更新');
      await loadData();
    } catch {
      Message.error('更新出站失败');
    } finally {
      setTogglingKey(null);
    }
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
            nodeGroups={nodeGroups}
            deletingKey={deletingKey}
            togglingKey={togglingKey}
            onEdit={handleEditRs}
            onDelete={handleDeleteRs}
            onChangeOutbound={handleChangeOutbound}
            onCopyURL={setCopyTarget}
          />
        </DataState>
      </div>

      <RuleSetModal visible={rsVisible} title={rsTitle} initialValues={rsForm} nodeGroups={nodeGroups}
        confirmLoading={submitting}
        onOk={handleRsOk} onCancel={() => setRsVisible(false)} />

      <RuleSetCopyURLModal
        visible={copyTarget !== null}
        ruleSet={copyTarget}
        devices={devices}
        systemHost={systemHost}
        onCancel={() => setCopyTarget(null)}
      />
    </>
  );
}
