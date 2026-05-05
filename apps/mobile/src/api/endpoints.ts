import { apiFetch } from '@/api/client';
import type {
  AuthResponse,
  UserProfile,
  Channel,
  ChannelMember,
  Message,
} from '@/types/shared';
import type { CreateChannelInput } from '@/utils/validation';

// Auth
export const authApi = {
  login: (email: string, password: string) =>
    apiFetch<AuthResponse>('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
      authRequired: false,
    }),

  register: (email: string, password: string) =>
    apiFetch<AuthResponse>('/api/v1/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
      authRequired: false,
    }),

  refresh: (refreshToken: string) =>
    apiFetch<AuthResponse>('/api/v1/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: refreshToken }),
      authRequired: false,
    }),

  logout: () =>
    apiFetch<void>('/api/v1/auth/logout', {
      method: 'POST',
      authRequired: true,
    }),

  me: () =>
    apiFetch<{ valid: boolean; user_id: string; email: string }>('/api/v1/auth/me'),
};

// Users
export const usersApi = {
  getProfile: () => apiFetch<UserProfile>('/api/v1/users/me'),

  updateProfile: (data: Partial<UserProfile>) =>
    apiFetch<UserProfile>('/api/v1/users/me', {
      method: 'PATCH',
      body: JSON.stringify(data),
    }),

  getById: (id: string) =>
    apiFetch<UserProfile>(`/api/v1/users/${id}`, { authRequired: false }),

  search: (q: string, limit = 20) =>
    apiFetch<UserProfile[]>(`/api/v1/users/search?q=${encodeURIComponent(q)}&limit=${limit}`),

  heartbeat: () =>
    apiFetch<void>('/api/v1/users/heartbeat', { method: 'POST' }),
};

// Channels
export const channelsApi = {
  listMine: () => apiFetch<Channel[]>('/api/v1/channels/me'),

  getById: (id: string) => apiFetch<Channel>(`/api/v1/channels/${id}`),

  create: (data: CreateChannelInput) =>
    apiFetch<Channel>('/api/v1/channels', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  createDm: (otherUserId: string) =>
    apiFetch<Channel>('/api/v1/channels/dm', {
      method: 'POST',
      body: JSON.stringify({ other_user_id: otherUserId }),
    }),

  delete: (id: string) =>
    apiFetch<{ success: boolean }>(`/api/v1/channels/${id}`, { method: 'DELETE' }),

  getMembers: (id: string) =>
    apiFetch<ChannelMember[]>(`/api/v1/channels/${id}/members`),

  addMember: (channelId: string, userId: string, role?: string) =>
    apiFetch<ChannelMember>(`/api/v1/channels/${channelId}/members`, {
      method: 'POST',
      body: JSON.stringify({ user_id: userId, ...(role ? { role } : {}) }),
    }),

  removeMember: (channelId: string, userId: string) =>
    apiFetch<{ success: boolean }>(
      `/api/v1/channels/${channelId}/members/${userId}`,
      { method: 'DELETE' },
    ),
};

// Messages
export const messagesApi = {
  list: (channelId: string, opts: { beforeId?: string; limit?: number } = {}) => {
    const params = new URLSearchParams({ channel_id: channelId, limit: String(opts.limit ?? 50) });
    if (opts.beforeId) params.set('before_id', opts.beforeId);
    return apiFetch<Message[]>(`/api/v1/messages?${params}`);
  },

  send: (channelId: string, content: string) =>
    apiFetch<Message>('/api/v1/messages', {
      method: 'POST',
      body: JSON.stringify({ channel_id: channelId, content }),
    }),
};

// Assist
export const assistApi = {
  suggestions: (channelId: string, currentInput: string, messageLimit = 10) =>
    apiFetch<{ suggestions: string[] }>('/api/v1/assist/suggestions', {
      method: 'POST',
      body: JSON.stringify({ channel_id: channelId, current_input: currentInput, message_limit: messageLimit }),
    }),
};
