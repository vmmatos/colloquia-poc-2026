import { useEffect, useRef } from 'react';
import { AppState } from 'react-native';
import * as LocalAuthentication from 'expo-local-authentication';
import { useAuthStore } from '@/auth/store';
import { useSettingsStore } from '@/stores/settings';
import { logger } from '@/utils/logger';

const LOCK_AFTER_MS = 5 * 60 * 1000; // 5 minutes

interface Options {
  onLock: () => void;
}

export function useAppForegroundLock({ onLock }: Options): void {
  const status = useAuthStore((s) => s.status);
  const biometricLockEnabled = useSettingsStore((s) => s.biometricLockEnabled);
  const backgroundAt = useRef<number | null>(null);

  useEffect(() => {
    if (status !== 'authenticated' || !biometricLockEnabled) return;

    const sub = AppState.addEventListener('change', async (state) => {
      if (state === 'background' || state === 'inactive') {
        backgroundAt.current = Date.now();
      } else if (state === 'active') {
        if (backgroundAt.current !== null) {
          const elapsed = Date.now() - backgroundAt.current;
          backgroundAt.current = null;
          if (elapsed >= LOCK_AFTER_MS) {
            // Verify hardware is still enrolled before locking.
            const hasHw = await LocalAuthentication.hasHardwareAsync();
            const enrolled = await LocalAuthentication.isEnrolledAsync();
            if (hasHw && enrolled) {
              logger.info('useAppForegroundLock: locking app');
              onLock();
            }
          }
        }
      }
    });

    return () => sub.remove();
  }, [status, biometricLockEnabled, onLock]);
}
