export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const authorization = getHeader(event, 'Authorization')
  const id = getRouterParam(event, 'id')
  const userId = getRouterParam(event, 'userId')

  return await $fetch(`${config.apiBase}/api/v1/channels/${id}/members/${userId}`, {
    method: 'DELETE',
    headers: { Authorization: authorization ?? '' },
  })
})
