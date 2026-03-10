import type { AuthState, AuthResponse, UserProfile, ValidateTokenResponse } from '../../shared/types/auth'

export function useAuth() {
  const config = useRuntimeConfig()
  const auth = useState<AuthState>('auth', () => ({
    user_id: null,
    access_token: null,
    expires_at: null,
  }))

  const isAuthenticated = computed(() => !!auth.value.access_token)

  const tokenExpiresIn = computed(() => {
    if (!auth.value.expires_at) return null
    const diff = new Date(auth.value.expires_at).getTime() - Date.now()
    if (diff <= 0) return 'expired'
    const secs = Math.floor(diff / 1000)
    return secs < 60 ? `${secs}s` : `${Math.floor(secs / 60)}m ${secs % 60}s`
  })

  async function register(email: string, password: string): Promise<AuthResponse> {
    const data = await $fetch<AuthResponse>('/api/auth/register', {
      method: 'POST',
      body: { email, password },
    })
    auth.value = { user_id: data.user_id, access_token: data.access_token, expires_at: data.expires_at }
    return data
  }

  async function login(email: string, password: string): Promise<AuthResponse> {
    const data = await $fetch<AuthResponse>('/api/auth/login', {
      method: 'POST',
      body: { email, password },
    })
    auth.value = { user_id: data.user_id, access_token: data.access_token, expires_at: data.expires_at }
    return data
  }

  async function logout(): Promise<void> {
    await $fetch('/api/auth/logout', {
      method: 'POST',
      body: { access_token: auth.value.access_token },
    })
    auth.value = { user_id: null, access_token: null, expires_at: null }
  }

  async function refreshToken(): Promise<AuthResponse> {
    const data = await $fetch<AuthResponse>('/api/auth/refresh', {
      method: 'POST',
    })
    auth.value = { user_id: data.user_id, access_token: data.access_token, expires_at: data.expires_at }
    return data
  }

  async function getProfile(): Promise<UserProfile> {
    return await $fetch<UserProfile>(`${config.public.apiBase}/api/v1/users/me`, {
      headers: { Authorization: `Bearer ${auth.value.access_token}` },
    })
  }

  async function validateToken(): Promise<ValidateTokenResponse> {
    return await $fetch<ValidateTokenResponse>(`${config.public.apiBase}/api/v1/auth/me`, {
      headers: { Authorization: `Bearer ${auth.value.access_token}` },
    })
  }

  return {
    auth,
    isAuthenticated,
    tokenExpiresIn,
    register,
    login,
    logout,
    refreshToken,
    getProfile,
    validateToken,
  }
}
