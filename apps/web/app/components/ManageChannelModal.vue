<script setup lang="ts">
import type { Channel, ChannelMember } from '../../shared/types/channels'
import type { UserProfile } from '../../shared/types/auth'

const props = defineProps<{
  open: boolean
  channelId: string
}>()

const emit = defineEmits<{
  close: []
  deleted: []
}>()

const { t } = useI18n()
const { auth } = useAuth()
const { fetchMembers, addMember, removeMember, deleteChannel } = useChannels()
const config = useRuntimeConfig()

const channel = ref<Channel | null>(null)
const members = ref<ChannelMember[]>([])
const profileMap = ref<Record<string, UserProfile>>({})
const myRole = ref<string>('')
const activeTab = ref<'members' | 'settings'>('members')
const loadError = ref('')

// Add member form
const memberSearch = ref('')
const userResults = ref<UserProfile[]>([])
const allUsers = ref<UserProfile[]>([])
const userSearchLoading = ref(false)
let userSearchTimer: ReturnType<typeof setTimeout> | null = null
const addLoading = ref(false)
const addError = ref('')

// Remove
const removingUserId = ref<string | null>(null)

// Delete
const confirmDelete = ref(false)
const deleteLoading = ref(false)
const deleteError = ref('')

async function loadData() {
  if (!props.channelId) return
  loadError.value = ''
  members.value = []
  profileMap.value = {}
  channel.value = null
  myRole.value = ''
  activeTab.value = 'members'
  memberSearch.value = ''
  userResults.value = []
  allUsers.value = []
  addError.value = ''
  confirmDelete.value = false

  try {
    const [ch, memberList] = await Promise.all([
      $fetch<Channel>(`/api/channels/${props.channelId}`, {
        headers: { Authorization: `Bearer ${auth.value.access_token}` },
      }),
      fetchMembers(props.channelId),
    ])
    channel.value = ch
    members.value = memberList
    const me = memberList.find(m => m.user_id === auth.value.user_id)
    myRole.value = me?.role ?? 'member'

    const profiles = await Promise.all(
      memberList.map(m =>
        $fetch<UserProfile>(`${config.public.apiBase}/api/v1/users/${m.user_id}`, {
          headers: { Authorization: `Bearer ${auth.value.access_token}` },
        }).catch(() => null)
      )
    )
    profileMap.value = Object.fromEntries(
      profiles.flatMap((p, i) => (p && memberList[i]) ? [[memberList[i]!.user_id, p]] : [])
    )

    // fetch all users for autocomplete (fire-and-forget)
    $fetch<UserProfile[]>(`${config.public.apiBase}/api/v1/users`, {
      headers: { Authorization: `Bearer ${auth.value.access_token}` },
    }).then(res => { allUsers.value = res ?? [] }).catch(() => {})
  } catch {
    loadError.value = t('channel.manage.loadError')
  }
}

const memberIds = computed(() => new Set(members.value.map(m => m.user_id)))

const displayedUsers = computed(() => {
  const base = memberSearch.value.length >= 3 ? userResults.value : allUsers.value
  return base.filter(u => !memberIds.value.has(u.user_id))
})

watch(memberSearch, (query) => {
  if (userSearchTimer) clearTimeout(userSearchTimer)
  if (query.length === 0) {
    userResults.value = []
    return
  }
  if (query.length < 3) {
    userResults.value = []
    return
  }
  userSearchTimer = setTimeout(async () => {
    userSearchLoading.value = true
    try {
      const res = await $fetch<UserProfile[]>(`${config.public.apiBase}/api/v1/users/search`, {
        query: { q: query },
        headers: { Authorization: `Bearer ${auth.value.access_token}` },
      })
      userResults.value = res ?? []
    } catch {
      userResults.value = []
    } finally {
      userSearchLoading.value = false
    }
  }, 300)
})

watch(() => props.open, (val) => {
  if (val) loadData()
}, { immediate: true })

watch(() => props.channelId, () => {
  if (props.open) loadData()
})

function close() {
  emit('close')
}

async function doAddUser(user: UserProfile) {
  addError.value = ''
  addLoading.value = true
  try {
    const newMember = await addMember(props.channelId, { user_id: user.user_id })
    members.value = [...members.value, newMember]
    profileMap.value[user.user_id] = user
    memberSearch.value = ''
    userResults.value = []
  } catch (err: unknown) {
    const msg = (err as { data?: { error?: string } })?.data?.error
    addError.value = msg ?? t('channel.manage.addError')
  } finally {
    addLoading.value = false
  }
}

async function doRemoveMember(userId: string) {
  removingUserId.value = userId
  try {
    await removeMember(props.channelId, userId)
    members.value = members.value.filter(m => m.user_id !== userId)
  } catch {
    // silently fail — refresh will show current state
  } finally {
    removingUserId.value = null
  }
}

async function doDelete() {
  deleteError.value = ''
  deleteLoading.value = true
  try {
    await deleteChannel(props.channelId)
    emit('deleted')
    close()
  } catch (err: unknown) {
    const msg = (err as { data?: { error?: string } })?.data?.error
    deleteError.value = msg ?? t('channel.manage.deleteError')
  } finally {
    deleteLoading.value = false
  }
}

function roleBadge(role: string) {
  if (role === 'owner') return { label: 'owner', cls: 'bg-primary/20 text-primary' }
  if (role === 'admin') return { label: 'admin', cls: 'bg-secondary text-foreground' }
  return { label: t('channel.manage.roleMember'), cls: 'bg-secondary/50 text-muted-foreground' }
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') close()
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <Teleport to="body">
    <Transition name="fade">
      <div
        v-if="open"
        class="fixed inset-0 z-50 flex items-center justify-center p-4"
      >
        <!-- Backdrop -->
        <div
          class="absolute inset-0 bg-background/60 backdrop-blur-sm"
          @click="close"
        />

        <!-- Modal -->
        <div class="relative z-10 w-full max-w-lg bg-card border border-border rounded-lg shadow-xl flex flex-col max-h-[85vh]">
          <!-- Header -->
          <div class="flex items-center justify-between px-6 py-4 border-b border-border flex-shrink-0">
            <div>
              <h2 class="font-heading font-semibold text-foreground text-base">
                # {{ channel?.name ?? '...' }}
              </h2>
              <p v-if="channel?.description" class="text-xs font-body text-muted-foreground mt-0.5">{{ channel.description }}</p>
            </div>
            <button
              class="text-muted-foreground hover:text-foreground transition-colors"
              @click="close"
            >
              <svg class="h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- Tabs -->
          <div class="flex border-b border-border flex-shrink-0">
            <button
              :class="[
                'px-6 py-2.5 text-sm font-heading transition-colors',
                activeTab === 'members'
                  ? 'text-foreground border-b-2 border-primary -mb-px'
                  : 'text-muted-foreground hover:text-foreground',
              ]"
              @click="activeTab = 'members'"
            >
              {{ $t('channel.manage.membersTab') }} <span class="text-xs text-muted-foreground ml-1">({{ members.length }})</span>
            </button>
            <button
              v-if="myRole === 'owner'"
              :class="[
                'px-6 py-2.5 text-sm font-heading transition-colors',
                activeTab === 'settings'
                  ? 'text-foreground border-b-2 border-primary -mb-px'
                  : 'text-muted-foreground hover:text-foreground',
              ]"
              @click="activeTab = 'settings'"
            >
              {{ $t('channel.manage.settingsTab') }}
            </button>
          </div>

          <!-- Error -->
          <div v-if="loadError" class="px-6 py-8 text-center">
            <p class="text-sm font-body text-muted-foreground">{{ loadError }}</p>
          </div>

          <!-- Members tab -->
          <div v-else-if="activeTab === 'members'" class="flex-1 overflow-y-auto">
            <!-- Member list -->
            <ul class="px-4 py-3 space-y-1">
              <li
                v-for="m in members"
                :key="m.user_id"
                class="flex items-center gap-3 py-1.5 px-2 rounded-md hover:bg-secondary/30 group"
              >
                <UiAvatar :name="profileMap[m.user_id]?.name || m.user_id" size="sm" />
                <div class="flex-1 min-w-0">
                  <p class="text-sm font-heading text-foreground truncate">{{ profileMap[m.user_id]?.name || m.user_id }}</p>
                </div>
                <button
                  v-if="(myRole === 'owner' || myRole === 'admin') && m.user_id !== auth.user_id && m.role !== 'owner'"
                  :disabled="removingUserId === m.user_id"
                  class="text-xs font-heading text-muted-foreground hover:text-destructive transition-colors opacity-0 group-hover:opacity-100 disabled:opacity-30"
                  @click="doRemoveMember(m.user_id)"
                >
                  {{ $t('channel.manage.remove') }}
                </button>
                <span :class="['text-xs px-2 py-0.5 rounded-full font-heading', roleBadge(m.role).cls]">
                  {{ roleBadge(m.role).label }}
                </span>
              </li>
            </ul>

            <!-- Add member section -->
            <div v-if="myRole === 'owner' || myRole === 'admin'" class="border-t border-border px-4 py-4">
              <p class="text-xs font-heading font-medium text-muted-foreground uppercase tracking-wider mb-3">
                {{ $t('channel.manage.addMember') }}
              </p>
              <div class="relative">
                <input
                  v-model="memberSearch"
                  type="text"
                  :placeholder="$t('channel.manage.searchPlaceholder')"
                  class="w-full bg-background border border-border rounded-md px-3 py-2 text-sm font-heading text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary/50 transition-colors"
                />
                <!-- Dropdown -->
                <ul
                  v-if="displayedUsers.length > 0"
                  class="absolute z-10 mt-1 w-full bg-card border border-border rounded-md shadow-lg max-h-48 overflow-y-auto"
                >
                  <li
                    v-for="u in displayedUsers"
                    :key="u.user_id"
                    class="flex items-center gap-2 px-3 py-2 cursor-pointer hover:bg-secondary/50 text-sm font-heading text-foreground"
                    @mousedown.prevent="doAddUser(u)"
                  >
                    <UiAvatar :name="u.name || u.user_id" size="sm" />
                    <span class="flex-1 truncate">{{ u.name || u.user_id }}</span>
                  </li>
                </ul>
                <!-- hint when query is 1-2 chars -->
                <p
                  v-else-if="memberSearch.length > 0 && memberSearch.length < 3"
                  class="text-xs text-muted-foreground font-body mt-1.5"
                >
                  {{ $t('channel.manage.searchHint', { n: 3 - memberSearch.length }) }}
                </p>
              </div>
              <p v-if="addError" class="text-xs text-destructive font-body mt-1.5">{{ addError }}</p>
            </div>
          </div>

          <!-- Settings tab (owner only) -->
          <div v-else-if="activeTab === 'settings'" class="flex-1 overflow-y-auto px-6 py-5">
            <div class="border border-destructive/30 rounded-lg p-5">
              <h3 class="font-heading font-semibold text-foreground text-sm mb-1">{{ $t('channel.manage.deleteTitle') }}</h3>
              <p class="text-xs font-body text-muted-foreground mb-4">
                {{ $t('channel.manage.deleteWarning') }}
              </p>

              <div v-if="!confirmDelete">
                <button
                  class="px-4 py-2 text-sm font-heading text-destructive border border-destructive/40 rounded-md hover:bg-destructive/10 transition-colors"
                  @click="confirmDelete = true"
                >
                  {{ $t('channel.manage.deleteButton') }}
                </button>
              </div>

              <div v-else class="space-y-3">
                <p class="text-sm font-heading text-foreground">
                  {{ $t('channel.manage.deleteConfirm') }}
                </p>
                <p class="font-mono text-sm text-primary">{{ channel?.name }}</p>
                <div class="flex gap-3">
                  <UiButton variant="danger" :loading="deleteLoading" @click="doDelete">
                    {{ $t('channel.manage.deleteSubmit') }}
                  </UiButton>
                  <button
                    class="px-4 py-2 text-sm font-heading text-muted-foreground hover:text-foreground transition-colors"
                    @click="confirmDelete = false"
                  >
                    {{ $t('common.cancel') }}
                  </button>
                </div>
                <p v-if="deleteError" class="text-xs text-destructive font-body">{{ deleteError }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
