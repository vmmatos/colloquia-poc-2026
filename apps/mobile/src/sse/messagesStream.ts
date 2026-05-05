import EventSource from 'react-native-sse';
import { AppState, AppStateStatus } from 'react-native';
import { getApiBaseUrl } from '@/utils/env';
import { tokenRotatedEmitter, calcBackoffMs } from '@/sse/reconnect';
import { logger } from '@/utils/logger';
import type { Message } from '@/types/shared';

type MessageHandler = (msg: Message) => void;

let es: InstanceType<typeof EventSource> | null = null;
let currentChannelIds: string[] = [];
let currentToken: string | null = null;
let retryCount = 0;
let retryTimer: ReturnType<typeof setTimeout> | null = null;
let handlers: Set<MessageHandler> = new Set();
let appStateSub: ReturnType<typeof AppState.addEventListener> | null = null;
let tokenSub: (() => void) | null = null;
let active = false;

function buildUrl(channelIds: string[], token: string): string {
  const base = getApiBaseUrl();
  const params = new URLSearchParams();
  channelIds.forEach((id) => params.append('channel_id', id));
  params.set('token', token);
  return `${base}/api/v1/messages/stream?${params}`;
}

function openStream(): void {
  if (!currentToken || currentChannelIds.length === 0) return;
  closeStream();

  const url = buildUrl(currentChannelIds, currentToken);
  logger.sse('messagesStream: opening', url);

  es = new EventSource(url);

  es.addEventListener('open', () => {
    retryCount = 0;
    logger.info('messagesStream: connected');
  });

  es.addEventListener('message', (event) => {
    if (!event.data) return;
    try {
      const msg = JSON.parse(event.data) as Message;
      handlers.forEach((h) => h(msg));
    } catch {
      // ignore malformed
    }
  });

  es.addEventListener('error', (event) => {
    logger.warn(`messagesStream: error, retry #${retryCount + 1}`, event);
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

export const messagesStream = {
  subscribe(onMessage: MessageHandler): () => void {
    handlers.add(onMessage);
    return () => handlers.delete(onMessage);
  },

  open(channelIds: string[], token: string): void {
    active = true;
    currentChannelIds = channelIds;
    currentToken = token;
    retryCount = 0;
    openStream();

    if (!appStateSub) {
      appStateSub = AppState.addEventListener('change', handleAppState);
    }

    if (!tokenSub) {
      tokenSub = tokenRotatedEmitter.on((newToken) => {
        if (newToken === null) {
          messagesStream.close();
        } else {
          currentToken = newToken;
          retryCount = 0;
          openStream();
        }
      });
    }
  },

  updateChannels(channelIds: string[]): void {
    if (
      JSON.stringify(channelIds.slice().sort()) ===
      JSON.stringify(currentChannelIds.slice().sort())
    ) return;
    currentChannelIds = channelIds;
    retryCount = 0;
    openStream();
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
