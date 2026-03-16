export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const authorization = getHeader(event, 'Authorization')
  const id = getRouterParam(event, 'id')

  return await $fetch(`${config.apiBase}/api/v1/channels/${id}`, {
    method: 'DELETE',
    headers: { Authorization: authorization ?? '' },
  })
})
