import { create } from 'zustand';
import { MMKV } from 'react-native-mmkv';
import type { Message, DraftMessage } from '@/types/shared';

const storage = new MMKV({ id: 'messages' });
const MAX_CACHE = 50;

function cacheKey(channelId: string): string {
  return `ch.${channelId}`;
}

function loadCached(channelId: string): Message[] {
  try {
    const raw = storage.getString(cacheKey(channelId));
    return raw ? (JSON.parse(raw) as Message[]) : [];
  } catch {
    return [];
  }
}

function saveToCache(channelId: string, msgs: Message[]): void {
  const toSave = msgs.slice(-MAX_CACHE);
  storage.set(cacheKey(channelId), JSON.stringify(toSave));
}

interface MessagesState {
  messages: Record<string, Message[]>;
  drafts: Record<string, DraftMessage[]>;

  getMessages: (channelId: string) => Message[];
  setHistory: (channelId: string, msgs: Message[]) => void;
  appendMessage: (channelId: string, msg: Message) => void;
  prependMessages: (channelId: string, msgs: Message[]) => void;
  replaceDraft: (channelId: string, tempId: string, msg: Message) => void;
  addDraft: (draft: DraftMessage) => void;
  updateDraftStatus: (tempId: string, status: DraftMessage['status']) => void;
  removeDraft: (tempId: string) => void;
  getDrafts: (channelId: string) => DraftMessage[];
}

export const useMessagesStore = create<MessagesState>((set, get) => ({
  messages: {},
  drafts: {},

  getMessages: (channelId) => get().messages[channelId] ?? loadCached(channelId),

  setHistory: (channelId, msgs) => {
    // Reverse from newest-first (API) to oldest-first (display).
    const sorted = [...msgs].reverse();
    saveToCache(channelId, sorted);
    set((s) => ({ messages: { ...s.messages, [channelId]: sorted } }));
  },

  appendMessage: (channelId, msg) => {
    set((s) => {
      const existing = s.messages[channelId] ?? loadCached(channelId);
      if (existing.some((m) => m.id === msg.id)) return s;
      const next = [...existing, msg];
      saveToCache(channelId, next);
      return { messages: { ...s.messages, [channelId]: next } };
    });
  },

  prependMessages: (channelId, msgs) => {
    set((s) => {
      const existing = s.messages[channelId] ?? loadCached(channelId);
      const existingIds = new Set(existing.map((m) => m.id));
      const newMsgs = [...msgs].reverse().filter((m) => !existingIds.has(m.id));
      const next = [...newMsgs, ...existing];
      saveToCache(channelId, next);
      return { messages: { ...s.messages, [channelId]: next } };
    });
  },

  replaceDraft: (channelId, tempId, msg) => {
    set((s) => {
      const channelDrafts = (s.drafts[channelId] ?? []).filter((d) => d.tempId !== tempId);
      const existing = s.messages[channelId] ?? loadCached(channelId);
      const next = existing.some((m) => m.id === msg.id) ? existing : [...existing, msg];
      saveToCache(channelId, next);
      return {
        messages: { ...s.messages, [channelId]: next },
        drafts: { ...s.drafts, [channelId]: channelDrafts },
      };
    });
  },

  addDraft: (draft) => {
    set((s) => {
      const prev = s.drafts[draft.channel_id] ?? [];
      return { drafts: { ...s.drafts, [draft.channel_id]: [...prev, draft] } };
    });
  },

  updateDraftStatus: (tempId, status) => {
    set((s) => {
      const updated: Record<string, DraftMessage[]> = {};
      for (const [ch, drafts] of Object.entries(s.drafts)) {
        updated[ch] = drafts.map((d) => (d.tempId === tempId ? { ...d, status } : d));
      }
      return { drafts: updated };
    });
  },

  removeDraft: (tempId) => {
    set((s) => {
      const updated: Record<string, DraftMessage[]> = {};
      for (const [ch, drafts] of Object.entries(s.drafts)) {
        updated[ch] = drafts.filter((d) => d.tempId !== tempId);
      }
      return { drafts: updated };
    });
  },

  getDrafts: (channelId) => get().drafts[channelId] ?? [],
}));
