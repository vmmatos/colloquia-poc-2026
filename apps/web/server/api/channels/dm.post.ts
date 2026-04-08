import type { Channel } from '../../../shared/types/channels'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const authorization = getHeader(event, 'Authorization')
  const body = await readBody<{ other_user_id: string }>(event)

  return await $fetch<Channel>(`${config.apiBase}/api/v1/channels/dm`, {
    method: 'POST',
    headers: { Authorization: authorization ?? '' },
    body,
  })
})
