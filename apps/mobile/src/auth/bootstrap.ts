import { authApi } from '@/api/endpoints';
import { getAccessTokenExpMs } from '@/auth/jwt';
import { configureClient } from '@/api/client';
import { useAuthStore } from '@/auth/store';
import { readRefreshToken, storeRefreshToken, deleteRefreshToken } from '@/auth/secureRefresh';
import { scheduleRefresh, cancelRefreshSchedule } from '@/auth/refreshScheduler';
import { tokenRotatedEmitter } from '@/sse/reconnect';
import { logger } from '@/utils/logger';

async function performRefresh(): Promise<void> {
  const refreshToken = await readRefreshToken();
  if (!refreshToken) {
    performLogout();
    throw new Error('No refresh token');
  }

  const data = await authApi.refresh(refreshToken);
  await storeRefreshToken(data.refresh_token);

  const expMs = getAccessTokenExpMs(data.access_token);
  useAuthStore.getState().setAuthenticated(data.user_id, data.access_token, expMs);
  scheduleRefresh(data.access_token, performRefresh);
  tokenRotatedEmitter.emit(data.access_token);
  logger.info('bootstrap: token refreshed');
}

function performLogout(): void {
  cancelRefreshSchedule();
  deleteRefreshToken();
  useAuthStore.getState().setUnauthenticated();
  tokenRotatedEmitter.emit(null);
}

export function setupClientAuth(): void {
  configureClient({ refresh: performRefresh, logout: performLogout });
}

// Called on app launch (before splash is hidden). No biometric prompt here —
// biometric is an optional app-lock, enabled separately in Profile settings.
export async function bootstrapAuth(): Promise<void> {
  setupClientAuth();

  const refreshToken = await readRefreshToken();
  if (!refreshToken) {
    useAuthStore.getState().setUnauthenticated();
    return;
  }

  try {
    const data = await authApi.refresh(refreshToken);
    await storeRefreshToken(data.refresh_token);
    const expMs = getAccessTokenExpMs(data.access_token);
    useAuthStore.getState().setAuthenticated(data.user_id, data.access_token, expMs);
    scheduleRefresh(data.access_token, performRefresh);
    tokenRotatedEmitter.emit(data.access_token);
    logger.info('bootstrap: authenticated');
  } catch (err) {
    logger.warn('bootstrap: refresh failed, clearing session', err);
    await deleteRefreshToken();
    useAuthStore.getState().setUnauthenticated();
  }
}

export async function handleAuthSuccess(data: {
  user_id: string;
  access_token: string;
  refresh_token: string;
}): Promise<void> {
  await storeRefreshToken(data.refresh_token);
  const expMs = getAccessTokenExpMs(data.access_token);
  useAuthStore.getState().setAuthenticated(data.user_id, data.access_token, expMs);
  scheduleRefresh(data.access_token, performRefresh);
  tokenRotatedEmitter.emit(data.access_token);
}

export async function handleLogout(): Promise<void> {
  try {
    await authApi.logout();
  } catch {
    // Best-effort: always proceed with local cleanup.
  }
  performLogout();
}

export { performRefresh };
