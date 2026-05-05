import { AppState, AppStateStatus } from 'react-native';
import { getAccessTokenExpMs } from '@/auth/jwt';
import { logger } from '@/utils/logger';

type RefreshFn = () => Promise<void>;

let timer: ReturnType<typeof setTimeout> | null = null;
let refreshFn: RefreshFn | null = null;

const BUFFER_MS = 60_000; // refresh 60s before expiry
const MIN_REFRESH_MS = 5_000;

export function scheduleRefresh(accessToken: string, refresh: RefreshFn): void {
  if (timer) clearTimeout(timer);
  refreshFn = refresh;

  let expMs: number;
  try {
    expMs = getAccessTokenExpMs(accessToken);
  } catch {
    logger.warn('refreshScheduler: could not decode JWT, using 14min fallback');
    expMs = Date.now() + 14 * 60_000;
  }

  const delay = Math.max(MIN_REFRESH_MS, expMs - Date.now() - BUFFER_MS);
  logger.info(`refreshScheduler: scheduled in ${Math.round(delay / 1000)}s`);
  timer = setTimeout(async () => {
    logger.info('refreshScheduler: executing refresh');
    await refresh();
  }, delay);
}

export function cancelRefreshSchedule(): void {
  if (timer) {
    clearTimeout(timer);
    timer = null;
  }
  refreshFn = null;
}

// Force-refresh if access token is within 90s of expiry when app becomes active.
export function setupForegroundRefresh(
  getExpMs: () => number | null,
  refresh: RefreshFn,
): () => void {
  const handler = (state: AppStateStatus) => {
    if (state === 'active') {
      const expMs = getExpMs();
      if (expMs !== null && expMs - Date.now() < 90_000) {
        logger.info('refreshScheduler: foregrounded with near-expiry, force refresh');
        void refresh();
      }
    }
  };
  const sub = AppState.addEventListener('change', handler);
  return () => sub.remove();
}
