export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const authorization = getHeader(event, 'Authorization')
  const body = await readBody(event)

  return await $fetch(`${config.apiBase}/api/v1/users/me`, {
    method: 'PATCH',
    body,
    headers: { Authorization: authorization ?? '' },
  })
})
