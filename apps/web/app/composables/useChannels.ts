import type { Channel, ChannelMember, CreateChannelInput, AddMemberInput } from '../../shared/types/channels'

export function useChannels() {
  const { authFetch } = useAuthFetch()
  const channels = useState<Channel[]>('channels', () => [])

  async function fetchMyChannels(): Promise<void> {
    channels.value = await authFetch<Channel[]>('/api/channels')
  }

  async function createChannel(input: CreateChannelInput): Promise<Channel> {
    const ch = await authFetch<Channel>('/api/channels', {
      method: 'POST',
      body: input,
    })
    channels.value = [...channels.value, ch]
    return ch
  }

  async function createDM(otherUserId: string): Promise<Channel> {
    const ch = await authFetch<Channel>('/api/channels/dm', {
      method: 'POST',
      body: { other_user_id: otherUserId },
    })
    if (!channels.value.find(c => c.id === ch.id)) {
      channels.value = [...channels.value, ch]
    }
    return ch
  }

  async function deleteChannel(id: string): Promise<void> {
    await authFetch(`/api/channels/${id}`, { method: 'DELETE' })
    channels.value = channels.value.filter(ch => ch.id !== id)
  }

  async function fetchChannel(id: string): Promise<Channel> {
    return await authFetch<Channel>(`/api/channels/${id}`)
  }

  async function fetchMembers(channelId: string): Promise<ChannelMember[]> {
    return await authFetch<ChannelMember[]>(`/api/channels/${channelId}/members`)
  }

  async function addMember(channelId: string, input: AddMemberInput): Promise<ChannelMember> {
    return await authFetch<ChannelMember>(`/api/channels/${channelId}/members`, {
      method: 'POST',
      body: input,
    })
  }

  async function removeMember(channelId: string, userId: string): Promise<void> {
    await authFetch(`/api/channels/${channelId}/members/${userId}`, { method: 'DELETE' })
  }

  return {
    channels,
    fetchMyChannels,
    createChannel,
    createDM,
    deleteChannel,
    fetchChannel,
    fetchMembers,
    addMember,
    removeMember,
  }
}
