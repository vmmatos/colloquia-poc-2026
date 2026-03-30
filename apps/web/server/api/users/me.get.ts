export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const authorization = getHeader(event, 'Authorization')

  return await $fetch(`${config.apiBase}/api/v1/users/me`, {
    headers: { Authorization: authorization ?? '' },
  })
})
