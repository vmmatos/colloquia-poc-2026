export function useIsMobile() {
  const isMobile = ref(false)
  onMounted(() => {
    const mql = window.matchMedia('(max-width: 767px)')
    isMobile.value = mql.matches
    const onChange = (e: MediaQueryListEvent) => { isMobile.value = e.matches }
    mql.addEventListener('change', onChange)
    onUnmounted(() => mql.removeEventListener('change', onChange))
  })
  return isMobile
}
