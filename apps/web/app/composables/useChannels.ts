import type { Channel, ChannelMember, CreateChannelInput, AddMemberInput } from '../../shared/types/channels'

export function useChannels() {
  const { auth } = useAuth()
  const channels = useState<Channel[]>('channels', () => [])

  function authHeaders() {
    return { Authorization: `Bearer ${auth.value.access_token}` }
  }

  async function fetchMyChannels(): Promise<void> {
    channels.value = await $fetch<Channel[]>('/api/channels', {
      headers: authHeaders(),
    })
  }

  async function createChannel(input: CreateChannelInput): Promise<Channel> {
    const ch = await $fetch<Channel>('/api/channels', {
      method: 'POST',
      headers: authHeaders(),
      body: input,
    })
    channels.value = [...channels.value, ch]
    return ch
  }

  async function createDM(otherUserId: string): Promise<Channel> {
    const ch = await $fetch<Channel>('/api/channels/dm', {
      method: 'POST',
      headers: authHeaders(),
      body: { other_user_id: otherUserId },
    })
    if (!channels.value.find(c => c.id === ch.id)) {
      channels.value = [...channels.value, ch]
    }
    return ch
  }

  async function deleteChannel(id: string): Promise<void> {
    await $fetch(`/api/channels/${id}`, {
      method: 'DELETE',
      headers: authHeaders(),
    })
    channels.value = channels.value.filter(ch => ch.id !== id)
  }

  async function fetchChannel(id: string): Promise<Channel> {
    return await $fetch<Channel>(`/api/channels/${id}`, {
      headers: authHeaders(),
    })
  }

  async function fetchMembers(channelId: string): Promise<ChannelMember[]> {
    return await $fetch<ChannelMember[]>(`/api/channels/${channelId}/members`, {
      headers: authHeaders(),
    })
  }

  async function addMember(channelId: string, input: AddMemberInput): Promise<ChannelMember> {
    return await $fetch<ChannelMember>(`/api/channels/${channelId}/members`, {
      method: 'POST',
      headers: authHeaders(),
      body: input,
    })
  }

  async function removeMember(channelId: string, userId: string): Promise<void> {
    await $fetch(`/api/channels/${channelId}/members/${userId}`, {
      method: 'DELETE',
      headers: authHeaders(),
    })
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
