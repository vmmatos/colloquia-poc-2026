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
  id: string
  user_id: string
  display_name: string
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
