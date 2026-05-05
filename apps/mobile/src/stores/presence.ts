import { create } from 'zustand';
import type { PresenceEvent } from '@/types/shared';

interface PresenceEntry {
  online: boolean;
  lastSeen: number;
}

interface PresenceState {
  presence: Record<string, PresenceEntry>;
  update: (event: PresenceEvent) => void;
  isOnline: (userId: string) => boolean;
}

export const usePresenceStore = create<PresenceState>((set, get) => ({
  presence: {},

  update: ({ user_id, online, last_seen }) => {
    set((s) => ({
      presence: { ...s.presence, [user_id]: { online, lastSeen: last_seen } },
    }));
  },

  isOnline: (userId) => get().presence[userId]?.online ?? false,
}));
