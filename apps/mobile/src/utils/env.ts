import { MMKV } from 'react-native-mmkv';

const devStorage = new MMKV({ id: 'dev-config' });

const envBase = process.env.EXPO_PUBLIC_API_BASE_URL ?? 'http://10.0.2.2';

export function getApiBaseUrl(): string {
  const override = devStorage.getString('dev.apiBaseUrl');
  if (override) return override.replace(/\/$/, '');
  return envBase.replace(/\/$/, '');
}

export function setDevApiBaseUrl(url: string): void {
  devStorage.set('dev.apiBaseUrl', url.replace(/\/$/, ''));
}

export function clearDevApiBaseUrl(): void {
  devStorage.delete('dev.apiBaseUrl');
}
