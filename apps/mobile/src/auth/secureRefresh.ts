import * as SecureStore from 'expo-secure-store';
import { logger } from '@/utils/logger';

const KEY = 'colloquia.refresh_token';

// Biometric is an app-lock feature, not a read gate.
// Tokens are stored with device-level encryption only.
export async function storeRefreshToken(token: string): Promise<void> {
  await SecureStore.setItemAsync(KEY, token, {
    keychainAccessible: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
  });
}

export async function readRefreshToken(): Promise<string | null> {
  try {
    return await SecureStore.getItemAsync(KEY);
  } catch (err) {
    logger.error('secureRefresh.readRefreshToken', err);
    return null;
  }
}

export async function deleteRefreshToken(): Promise<void> {
  try {
    await SecureStore.deleteItemAsync(KEY);
  } catch {
    // ignore
  }
}
