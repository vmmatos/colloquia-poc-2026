import type { AuthResponse } from '../../../shared/types/auth'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const body = await readBody(event)

  const upstream = await $fetch<AuthResponse & { refresh_token: string }>(
    `${config.apiBase}/api/v1/auth/register`,
    {
      method: 'POST',
      body,
    }
  )

  setCookie(event, 'refresh_token', upstream.refresh_token, {
    httpOnly: true,
    sameSite: 'strict',
    secure: !import.meta.dev,
    path: '/',
    maxAge: 60 * 60 * 24 * 30, // 30 days
  })

  console.log(`[AUTH] Register success user_id=${upstream.user_id}`)

  const { refresh_token: _rt, ...response } = upstream
  return {
    ...response,
    expires_at: new Date(Number(response.expires_at) * 1000).toISOString(),
  } as AuthResponse
})
