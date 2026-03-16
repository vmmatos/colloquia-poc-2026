import type { Channel, CreateChannelInput } from '../../../shared/types/channels'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const authorization = getHeader(event, 'Authorization')
  const body = await readBody<CreateChannelInput>(event)

  return await $fetch<Channel>(`${config.apiBase}/api/v1/channels`, {
    method: 'POST',
    headers: { Authorization: authorization ?? '' },
    body,
  })
})
