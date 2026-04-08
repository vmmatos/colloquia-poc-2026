export interface AuthResponse {
  user_id: string
  access_token: string
  expires_at: string
}

export interface AuthState {
  user_id: string | null
  access_token: string | null
  expires_at: string | null
}

export interface UserProfile {
  user_id: string
  name: string
  avatar_url: string
  bio: string
  created_at: string
  updated_at: string
}

export interface ValidateTokenResponse {
  valid: boolean
  user_id: string
  email: string
}

export interface PresenceEvent {
  user_id: string
  online: boolean
  last_seen: number // Unix seconds
}
