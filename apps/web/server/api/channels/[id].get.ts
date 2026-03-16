import type { Channel } from '../../../shared/types/channels'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const authorization = getHeader(event, 'Authorization')
  const id = getRouterParam(event, 'id')

  return await $fetch<Channel>(`${config.apiBase}/api/v1/channels/${id}`, {
    headers: { Authorization: authorization ?? '' },
  })
})
