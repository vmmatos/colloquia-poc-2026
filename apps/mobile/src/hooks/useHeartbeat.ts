import { useEffect, useRef } from 'react';
import { AppState } from 'react-native';
import { usersApi } from '@/api/endpoints';
import { useAuthStore } from '@/auth/store';
import { logger } from '@/utils/logger';

const INTERVAL_MS = 10_000;

export function useHeartbeat(): void {
  const status = useAuthStore((s) => s.status);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  function startHeartbeat() {
    if (intervalRef.current) return;
    intervalRef.current = setInterval(() => {
      usersApi.heartbeat().catch((e) => logger.warn('heartbeat failed', e));
    }, INTERVAL_MS);
  }

  function stopHeartbeat() {
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
  }

  useEffect(() => {
    if (status !== 'authenticated') {
      stopHeartbeat();
      return;
    }

    startHeartbeat();

    const sub = AppState.addEventListener('change', (state) => {
      if (state === 'active') startHeartbeat();
      else stopHeartbeat();
    });

    return () => {
      stopHeartbeat();
      sub.remove();
    };
  }, [status]);
}
