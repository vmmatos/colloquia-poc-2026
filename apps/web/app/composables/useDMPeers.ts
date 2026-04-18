export function useDMPeers() {
  const { auth } = useAuth()
  const peers = useState<Record<string, string>>('dmPeers', () => ({}))

  function setPeer(channelId: string, memberUserIds: string[]) {
    const other = memberUserIds.find(id => id !== auth.value.user_id)
    if (other) peers.value[channelId] = other
  }

  function getPeer(channelId: string): string | null {
    return peers.value[channelId] ?? null
  }

  return { peers, setPeer, getPeer }
}
