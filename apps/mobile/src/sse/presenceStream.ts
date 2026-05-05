import EventSource from 'react-native-sse';
import { AppState, AppStateStatus } from 'react-native';
import { getApiBaseUrl } from '@/utils/env';
import { tokenRotatedEmitter, calcBackoffMs } from '@/sse/reconnect';
import { logger } from '@/utils/logger';
import type { PresenceEvent } from '@/types/shared';

type PresenceHandler = (event: PresenceEvent) => void;

let es: InstanceType<typeof EventSource> | null = null;
let currentToken: string | null = null;
let retryCount = 0;
let retryTimer: ReturnType<typeof setTimeout> | null = null;
let handlers: Set<PresenceHandler> = new Set();
let appStateSub: ReturnType<typeof AppState.addEventListener> | null = null;
let tokenSub: (() => void) | null = null;
let active = false;

function buildUrl(token: string): string {
  const base = getApiBaseUrl();
  return `${base}/api/v1/users/presence/stream?token=${encodeURIComponent(token)}`;
}

function openStream(): void {
  if (!currentToken) return;
  closeStream();

  const url = buildUrl(currentToken);
  logger.sse('presenceStream: opening', url);

  es = new EventSource(url);

  es.addEventListener('open', () => {
    retryCount = 0;
    logger.info('presenceStream: connected');
  });

  es.addEventListener('message', (event) => {
    if (!event.data) return;
    try {
      const evt = JSON.parse(event.data) as PresenceEvent;
      handlers.forEach((h) => h(evt));
    } catch {
      // ignore
    }
  });

  es.addEventListener('error', () => {
    logger.warn(`presenceStream: error, retry #${retryCount + 1}`);
    closeStream();
    scheduleRetry();
  });
}

function closeStream(): void {
  if (es) {
    es.removeAllEventListeners();
    es.close();
    es = null;
  }
  if (retryTimer) {
    clearTimeout(retryTimer);
    retryTimer = null;
  }
}

function scheduleRetry(): void {
  if (!active) return;
  const delay = calcBackoffMs(retryCount++);
  retryTimer = setTimeout(openStream, delay);
}

function handleAppState(state: AppStateStatus): void {
  if (state === 'active' && active) {
    retryCount = 0;
    openStream();
  } else if (state === 'background') {
    closeStream();
  }
}

export const presenceStream = {
  subscribe(onEvent: PresenceHandler): () => void {
    handlers.add(onEvent);
    return () => handlers.delete(onEvent);
  },

  open(token: string): void {
    active = true;
    currentToken = token;
    retryCount = 0;
    openStream();

    if (!appStateSub) {
      appStateSub = AppState.addEventListener('change', handleAppState);
    }

    if (!tokenSub) {
      tokenSub = tokenRotatedEmitter.on((newToken) => {
        if (newToken === null) {
          presenceStream.close();
        } else {
          currentToken = newToken;
          retryCount = 0;
          openStream();
        }
      });
    }
  },

  close(): void {
    active = false;
    closeStream();
    handlers.clear();
    appStateSub?.remove();
    appStateSub = null;
    tokenSub?.();
    tokenSub = null;
  },
};
