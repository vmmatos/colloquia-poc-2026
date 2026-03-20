export default defineNuxtRouteMiddleware(() => {
  const authReady = useState<boolean>('authReady', () => false)
  const { isAuthenticated } = useAuth()
  if (!authReady.value) return
  if (!isAuthenticated.value) return navigateTo('/login')
})
