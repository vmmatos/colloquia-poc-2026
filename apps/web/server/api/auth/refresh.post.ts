import type { AuthResponse } from '../../../shared/types/auth'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const refreshToken = getCookie(event, 'refresh_token')

  if (!refreshToken) {
    throw createError({ statusCode: 401, message: 'No refresh token' })
  }

  let upstream: AuthResponse & { refresh_token: string }

  try {
    upstream = await $fetch<AuthResponse & { refresh_token: string }>(
      `${config.apiBase}/api/v1/auth/refresh`,
      {
        method: 'POST',
        body: { refresh_token: refreshToken },
      }
    )
  } catch (err: unknown) {
    const statusCode = (err as { statusCode?: number }).statusCode
    if (statusCode === 401 || statusCode === 403) {
      deleteCookie(event, 'refresh_token', { path: '/' })
    }
    throw err
  }

  setCookie(event, 'refresh_token', upstream.refresh_token, {
    httpOnly: true,
    sameSite: 'strict',
    secure: !import.meta.dev,
    path: '/',
    maxAge: 60 * 60 * 24 * 30,
  })

  console.log(`[AUTH] Token refreshed user_id=${upstream.user_id}`)

  const { refresh_token: _rt, ...response } = upstream
  return response as AuthResponse
})
