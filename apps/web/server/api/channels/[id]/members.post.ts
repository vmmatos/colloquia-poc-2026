import type { ChannelMember, AddMemberInput } from '../../../../shared/types/channels'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const authorization = getHeader(event, 'Authorization')
  const id = getRouterParam(event, 'id')
  const body = await readBody<AddMemberInput>(event)

  return await $fetch<ChannelMember>(`${config.apiBase}/api/v1/channels/${id}/members`, {
    method: 'POST',
    headers: { Authorization: authorization ?? '' },
    body,
  })
})
