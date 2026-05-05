import { create } from 'zustand';
import { MMKV } from 'react-native-mmkv';
import { channelsApi } from '@/api/endpoints';
import type { Channel } from '@/types/shared';
import { logger } from '@/utils/logger';

const storage = new MMKV({ id: 'channels' });

function loadCached(): Channel[] {
  try {
    const raw = storage.getString('channels.mine');
    return raw ? (JSON.parse(raw) as Channel[]) : [];
  } catch {
    return [];
  }
}

function saveCached(channels: Channel[]): void {
  storage.set('channels.mine', JSON.stringify(channels));
}

interface ChannelsState {
  channels: Channel[];
  loading: boolean;
  error: string | null;
  unread: Record<string, number>;
  lastFetchAt: number | null;

  fetch: () => Promise<void>;
  upsertChannel: (ch: Channel) => void;
  incrementUnread: (channelId: string) => void;
  clearUnread: (channelId: string) => void;
  removeChannel: (id: string) => void;
}

export const useChannelsStore = create<ChannelsState>((set, get) => ({
  channels: loadCached(),
  loading: false,
  error: null,
  unread: {},
  lastFetchAt: null,

  fetch: async () => {
    set({ loading: true, error: null });
    try {
      const channels = await channelsApi.listMine();
      saveCached(channels);
      set({ channels, loading: false, lastFetchAt: Date.now() });
    } catch (err) {
      logger.error('channelsStore.fetch', err);
      set({ loading: false, error: 'Erro a carregar canais' });
    }
  },

  upsertChannel: (ch) => {
    set((s) => {
      const idx = s.channels.findIndex((c) => c.id === ch.id);
      const next =
        idx >= 0
          ? s.channels.map((c) => (c.id === ch.id ? ch : c))
          : [ch, ...s.channels];
      saveCached(next);
      return { channels: next };
    });
  },

  incrementUnread: (channelId) => {
    set((s) => ({ unread: { ...s.unread, [channelId]: (s.unread[channelId] ?? 0) + 1 } }));
  },

  clearUnread: (channelId) => {
    set((s) => {
      const next = { ...s.unread };
      delete next[channelId];
      return { unread: next };
    });
  },

  removeChannel: (id) => {
    set((s) => {
      const next = s.channels.filter((c) => c.id !== id);
      saveCached(next);
      return { channels: next };
    });
  },
}));
