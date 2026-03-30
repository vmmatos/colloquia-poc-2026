export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const authorization = getHeader(event, 'Authorization')
  const query = getQuery(event)

  return await $fetch(`${config.apiBase}/api/v1/messages`, {
    query,
    headers: { Authorization: authorization ?? '' },
  })
})
