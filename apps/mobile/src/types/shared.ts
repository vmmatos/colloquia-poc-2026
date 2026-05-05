// Espelho dos contratos do BE — adaptados para chamadas directas (sem BFF Nuxt).
// expires_at da resposta de auth é Unix seconds (session/refresh expiry, 7 dias).
// O expiry do access token (15 min) é extraído do JWT.exp claim.

export interface AuthResponse {
  user_id: string;
  access_token: string;
  refresh_token: string;
  expires_at: number; // Unix seconds (session expiry = refresh token expiry)
}

export interface UserProfile {
  user_id: string;
  email: string;
  name: string;
  avatar: string;
  bio: string;
  timezone: string;
  status: string;
  language?: string;
  created_at: number;
  updated_at: number;
}

export interface Channel {
  id: string;
  name: string;
  description: string;
  is_private: boolean;
  created_by: string;
  archived: boolean;
  type: 'dm' | 'group' | 'channel';
  dm_key?: string;
  member_count: number;
  created_at: number;
  updated_at: number;
}

export interface ChannelMember {
  channel_id: string;
  user_id: string;
  role: 'owner' | 'admin' | 'member';
  joined_at: number;
}

export interface Message {
  id: string;
  channel_id: string;
  user_id: string;
  content: string;
  created_at: number; // Unix seconds
}

export interface PresenceEvent {
  user_id: string;
  online: boolean;
  last_seen: number; // Unix seconds
}

export interface CreateChannelInput {
  name: string;
  description?: string;
  is_private: boolean;
  type: 'dm' | 'group' | 'channel';
  initial_member_ids?: string[];
}

export interface SearchUsersResult {
  results: UserProfile[];
}

export type DraftStatus = 'pending' | 'failed';

export interface DraftMessage {
  tempId: string;
  channel_id: string;
  content: string;
  status: DraftStatus;
  created_at: number; // local timestamp
}
