// Module-level state — shared across all composable instances (same pattern as useNotifications)
const peers = reactive<Record<string, string>>({}) // channelId → otherUserId

export function useDMPeers() {
  const { auth } = useAuth()

  function setPeer(channelId: string, memberUserIds: string[]) {
    const other = memberUserIds.find(id => id !== auth.value.user_id)
    if (other) peers[channelId] = other
  }

  function getPeer(channelId: string): string | null {
    return peers[channelId] ?? null
  }

  return { peers, setPeer, getPeer }
}
