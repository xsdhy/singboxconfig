import type { SingDNS } from '../types';

export const DNS_SETTING_KEY = 'dns_config';

// 与当前后端 convert/singbox/dns.go 的默认内容保持一致，用于“恢复默认”按钮。
export const DEFAULT_DNS_CONFIG: SingDNS = {
  servers: [
    {
      tag: 'dns_proxy',
      address: 'https://1.1.1.1/dns-query',
      address_resolver: 'dns_resolver',
      strategy: 'ipv4_only',
      detour: 'general',
    },
    {
      tag: 'dns_direct',
      address: 'h3://dns.alidns.com/dns-query',
      address_resolver: 'dns_resolver',
      strategy: 'ipv4_only',
      detour: 'direct',
    },
    {
      tag: 'dns_block',
      address: 'rcode://refused',
    },
    {
      tag: 'dns_resolver',
      address: '223.5.5.5',
      strategy: 'ipv4_only',
      detour: 'direct',
    },
  ],
  rules: [
    {
      outbound: 'any',
      server: 'dns_resolver',
    },
    {
      clash_mode: 'direct',
      server: 'dns_direct',
    },
    {
      clash_mode: 'global',
      server: 'dns_proxy',
    },
    {
      rule_set: 'cnip',
      server: 'dns_direct',
    },
    {
      rule_set: 'cnsite',
      server: 'dns_direct',
    },
  ],
  final: 'dns_proxy',
};

export function getDefaultDnsEditorText(): string {
  return JSON.stringify(DEFAULT_DNS_CONFIG, null, 2);
}

export function getDnsEditorText(settingValue?: string | null): string {
  if (!settingValue) {
    return getDefaultDnsEditorText();
  }

  try {
    return JSON.stringify(JSON.parse(settingValue), null, 2);
  } catch {
    return settingValue;
  }
}
