export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const body = await readBody(event)

  try {
    await $fetch(`${config.apiBase}/api/v1/auth/logout`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${body.access_token}`,
      },
    })
    console.log('[AUTH] Logout success')
  } catch {
    console.log('[AUTH] Logout upstream failed, clearing cookie anyway')
  }

  deleteCookie(event, 'refresh_token', { path: '/' })

  return { success: true }
})
