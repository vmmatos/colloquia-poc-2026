export interface Channel {
  id: string
  name: string
  description: string
  is_private: boolean
  created_by: string
  archived: boolean
  member_count: number
  created_at: number
  updated_at: number
}

export interface ChannelMember {
  channel_id: string
  user_id: string
  role: 'owner' | 'admin' | 'member'
  joined_at: number
}

export interface CreateChannelInput {
  name: string
  description?: string
  is_private: boolean
  initial_member_ids?: string[]
}

export interface AddMemberInput {
  user_id: string
  role?: 'admin' | 'member'
}
