// 系统 Host 在全局设置存储中的固定 key，与后端 service.systemHostSettingKey 保持一致。
export const SYSTEM_HOST_SETTING_KEY = 'system_host';

// 校验系统 Host 是否为合法的 http/https 绝对地址；空串视为未配置，允许。
// 与后端 service.validateSystemHost 保持一致，避免提交后被后端拒绝。
export function isValidSystemHost(raw: string): boolean {
  const host = raw.trim().replace(/\/+$/, '');
  if (host === '') {
    return true;
  }
  try {
    const parsed = new URL(host);
    return parsed.protocol === 'http:' || parsed.protocol === 'https:';
  } catch {
    return false;
  }
}

// 规范化系统 Host：去掉首尾空白与尾部斜杠，与后端规范化逻辑一致。
export function normalizeSystemHost(raw: string): string {
  return raw.trim().replace(/\/+$/, '');
}
