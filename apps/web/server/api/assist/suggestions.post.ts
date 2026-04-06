export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const authorization = getHeader(event, 'Authorization')
  const body = await readBody(event)

  return await $fetch(`${config.apiBase}/api/v1/assist/suggestions`, {
    method: 'POST',
    body,
    headers: { Authorization: authorization ?? '' },
  })
})
