export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const id = getRouterParam(event, 'id')
  return await $fetch(`${config.apiBase}/api/v1/users/${id}`)
})
