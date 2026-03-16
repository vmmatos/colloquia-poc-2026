import type { ChannelMember } from '../../../../shared/types/channels'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const authorization = getHeader(event, 'Authorization')
  const id = getRouterParam(event, 'id')

  return await $fetch<ChannelMember[]>(`${config.apiBase}/api/v1/channels/${id}/members`, {
    headers: { Authorization: authorization ?? '' },
  })
})
