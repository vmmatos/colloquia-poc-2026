<script setup lang="ts">
import type { Toast } from '~/components/MessageToast.vue'
import type { SseEvent } from '~/composables/useSSE'

const { auth, logout, getProfile } = useAuth()
useTokenRefresh()
const isMobile = useIsMobile()
const { addNotification } = useNotifications()
const { channels, fetchMyChannels, fetchMembers } = useChannels()
const { resolveUser, prefetchUsers } = useUsersCache()
const { getPeer, setPeer } = useDMPeers()

const displayName = ref('')
const showProfile = ref(false)
const sidebarOpen = ref(false)
const showCreateChannel = ref(false)
const showNewDM = ref(false)
const managingChannelId = ref<string | null>(null)

const channelsOpen = ref(true)
const dmsOpen = ref(true)
const toasts = ref<Toast[]>([])
const unreadCounts = reactive<Record<string, number>>({})

// ── Channel splits ─────────────────────────────────────────────────────────────

const regularChannels = computed(() => channels.value.filter(c => c.type === 'channel'))
const dmAndGroupChannels = computed(() => channels.value.filter(c => c.type === 'dm' || c.type === 'group'))

// ── SSE ────────────────────────────────────────────────────────────────────────

const route = useRoute()
const activeChannelId = computed(() => {
  const id = route.params.id
  return typeof id === 'string' ? id : null
})

const activeChannelHandler = ref<((e: SseEvent) => void) | null>(null)
provide('registerChannelHandler', (fn: ((e: SseEvent) => void) | null) => {
  activeChannelHandler.value = fn
})

function onSseMessage(event: SseEvent) {
  const isViewing = event.channel_id === activeChannelId.value

  if (isViewing && activeChannelHandler.value) {
    activeChannelHandler.value(event)
    return
  }

  // Canal em background → contagem de não lidos + notificação + toast
  unreadCounts[event.channel_id] = (unreadCounts[event.channel_id] ?? 0) + 1
  const ch = channels.value.find(c => c.id === event.channel_id)
  const channelLabel = dmChannelLabel(ch?.id ?? '', ch?.type ?? 'channel', ch?.name ?? event.channel_id)
  addNotification({
    type: 'message',
    title: `Nova mensagem em ${channelLabel}`,
    body: event.content.slice(0, 80),
    time: 'agora',
    channelId: event.channel_id,
  })

  const toastId = Date.now()
  toasts.value.push({
    id: toastId,
    author: resolveUser(event.user_id),
    preview: event.content.slice(0, 60),
    channel: channelLabel,
  })
  setTimeout(() => { toasts.value = toasts.value.filter(t => t.id !== toastId) }, 3500)
}

const sse = useSSE({ activeChannelId, onMessage: onSseMessage })

watch(activeChannelId, (id) => {
  if (id) unreadCounts[id] = 0
}, { immediate: true })

// ── DM peer resolution ─────────────────────────────────────────────────────────

async function loadDMPeers(chList = channels.value) {
  const dmChannels = chList.filter(c => c.type === 'dm')
  await Promise.all(dmChannels.map(async (ch) => {
    if (getPeer(ch.id)) return
    try {
      const members = await fetchMembers(ch.id)
      const ids = members.map(m => m.user_id)
      setPeer(ch.id, ids)
      await prefetchUsers(ids)
    } catch { /* best-effort */ }
  }))
}

function dmChannelLabel(channelId: string, type: string, name: string): string {
  if (type === 'dm') {
    const peerId = getPeer(channelId)
    return peerId ? resolveUser(peerId) : '...'
  }
  if (type === 'group') return name || 'Grupo'
  return `#${name}`
}

// ── Lifecycle ──────────────────────────────────────────────────────────────────

let channelPollTimer: ReturnType<typeof setInterval> | null = null

onMounted(async () => {
  try {
    const profile = await getProfile()
    displayName.value = profile.name || ''
  } catch {
    // profile might not exist yet
  }

  try {
    await fetchMyChannels()
    sse.subscribeToChannels(channels.value.map(c => c.id))
    loadDMPeers()
  } catch {
    // channels might not be available
  }

  channelPollTimer = setInterval(async () => {
    try {
      await fetchMyChannels()
      loadDMPeers()
    } catch { /* ignore */ }
  }, 30_000)
})

onUnmounted(() => {
  sse.closeAll()
  if (channelPollTimer) clearInterval(channelPollTimer)
})

watch(channels, (newChannels) => {
  sse.subscribeToChannels(newChannels.map(c => c.id))
  loadDMPeers(newChannels)
}, { deep: false })

watch(() => auth.value.access_token, (token) => {
  if (token && channels.value.length > 0) {
    sse.closeAll()
    nextTick(() => sse.subscribeToChannels(channels.value.map(c => c.id)))
  }
})

// ── Actions ────────────────────────────────────────────────────────────────────

async function handleLogout() {
  try { await logout() } catch { /* ignore */ }
  await navigateTo('/login')
}

// Provide panel state for pages that need it
provide('showProfile', showProfile)
provide('openSidebar', () => { sidebarOpen.value = true })

async function onProfileClose() {
  showProfile.value = false
  try {
    const profile = await getProfile()
    displayName.value = profile.name || ''
  } catch {
    // ignore
  }
}

function dismissToast(id: number) {
  toasts.value = toasts.value.filter(t => t.id !== id)
}
</script>

<template>
  <div class="flex h-screen overflow-hidden bg-background">
    <!-- Mobile backdrop -->
    <div
      v-if="isMobile && sidebarOpen"
      class="fixed inset-0 bg-background/60 backdrop-blur-sm z-30"
      @click="sidebarOpen = false"
    />

    <!-- Sidebar wrapper -->
    <div
      :class="[
        isMobile
          ? 'fixed inset-y-0 left-0 z-40 transform transition-transform duration-200 ease-out'
          : '',
        isMobile && !sidebarOpen ? '-translate-x-full' : '',
      ]"
    >
      <!-- Left Sidebar -->
      <aside class="w-60 flex-shrink-0 bg-sidebar flex flex-col h-screen border-r border-border">
        <!-- Workspace header -->
        <div class="h-14 flex items-center px-4 border-b border-border flex-shrink-0">
          <span class="text-foreground font-heading font-semibold text-base tracking-tight">Colloquia</span>
        </div>

        <!-- Scrollable nav -->
        <nav class="flex-1 overflow-y-auto py-3">
          <!-- Channels section -->
          <div class="mb-2">
            <div class="flex items-center w-full px-4 py-1">
              <button
                class="flex items-center gap-1 flex-1 text-xs uppercase tracking-wider text-muted-foreground hover:text-foreground transition-colors font-heading"
                @click="channelsOpen = !channelsOpen"
              >
                <svg
                  :class="['h-3 w-3 transition-transform flex-shrink-0', channelsOpen ? 'rotate-90' : '']"
                  xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor"
                >
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                </svg>
                <span>Canais</span>
              </button>
              <button
                class="text-muted-foreground hover:text-foreground transition-colors ml-1 flex-shrink-0"
                title="Criar canal"
                @click.stop="showCreateChannel = true"
              >
                <svg class="h-3.5 w-3.5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
                </svg>
              </button>
            </div>

            <ul v-if="channelsOpen" class="mt-1">
              <li
                v-for="ch in regularChannels"
                :key="ch.id"
                class="group relative flex items-stretch"
                :class="[
                  $route.params.id === ch.id ? 'bg-secondary text-foreground' : 'text-muted-foreground hover:bg-secondary/50 hover:text-foreground',
                ]"
              >
                <NuxtLink
                  :to="`/channels/${ch.id}`"
                  class="flex-1 flex items-center gap-2 px-4 py-1.5 text-sm font-heading transition-colors"
                >
                  <span class="text-muted-foreground">#</span>
                  <span class="flex-1 truncate">{{ ch.name }}</span>
                  <span
                    v-if="(unreadCounts[ch.id] ?? 0) > 0"
                    class="ml-auto w-4 h-4 rounded-full bg-primary text-primary-foreground text-xs font-heading font-semibold flex items-center justify-center flex-shrink-0"
                  >
                    {{ (unreadCounts[ch.id] ?? 0) > 9 ? '9+' : unreadCounts[ch.id] }}
                  </span>
                </NuxtLink>
                <button
                  class="opacity-0 group-hover:opacity-100 px-2 text-muted-foreground hover:text-foreground transition-opacity flex-shrink-0"
                  title="Gerir canal"
                  @click="managingChannelId = ch.id"
                >
                  <svg class="h-3 w-3" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  </svg>
                </button>
              </li>

              <li v-if="regularChannels.length === 0">
                <button
                  class="flex items-center gap-2 w-full px-4 py-1.5 text-sm font-heading text-muted-foreground/60 hover:text-muted-foreground transition-colors"
                  @click="showCreateChannel = true"
                >
                  <span class="text-muted-foreground/40">#</span>
                  <span class="italic">Criar primeiro canal...</span>
                </button>
              </li>
            </ul>
          </div>

          <!-- DMs & Groups section -->
          <div>
            <div class="flex items-center w-full px-4 py-1">
              <button
                class="flex items-center gap-1 flex-1 text-xs uppercase tracking-wider text-muted-foreground hover:text-foreground transition-colors font-heading"
                @click="dmsOpen = !dmsOpen"
              >
                <svg
                  :class="['h-3 w-3 transition-transform flex-shrink-0', dmsOpen ? 'rotate-90' : '']"
                  xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor"
                >
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                </svg>
                <span>Mensagens directas</span>
              </button>
              <button
                class="text-muted-foreground hover:text-foreground transition-colors ml-1 flex-shrink-0"
                title="Nova mensagem directa"
                @click.stop="showNewDM = true"
              >
                <svg class="h-3.5 w-3.5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
                </svg>
              </button>
            </div>

            <ul v-if="dmsOpen" class="mt-1">
              <li
                v-for="ch in dmAndGroupChannels"
                :key="ch.id"
                class="group relative flex items-stretch"
                :class="[
                  $route.params.id === ch.id ? 'bg-secondary text-foreground' : 'text-muted-foreground hover:bg-secondary/50 hover:text-foreground',
                ]"
              >
                <NuxtLink
                  :to="`/channels/${ch.id}`"
                  class="flex-1 flex items-center gap-2 px-4 py-1.5 text-sm font-heading transition-colors"
                  :class="(unreadCounts[ch.id] ?? 0) > 0 ? 'font-semibold' : ''"
                >
                  <!-- DM: avatar of the other person -->
                  <template v-if="ch.type === 'dm'">
                    <div class="flex-shrink-0">
                      <UiAvatar :name="getPeer(ch.id) ? resolveUser(getPeer(ch.id)!) : '?'" size="sm" />
                    </div>
                    <span class="flex-1 truncate">
                      {{ getPeer(ch.id) ? resolveUser(getPeer(ch.id)!) : '...' }}
                    </span>
                  </template>

                  <!-- Group: people icon + name -->
                  <template v-else>
                    <svg class="h-3.5 w-3.5 flex-shrink-0 text-muted-foreground" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
                    </svg>
                    <span class="flex-1 truncate">{{ ch.name || 'Grupo' }}</span>
                  </template>

                  <span
                    v-if="(unreadCounts[ch.id] ?? 0) > 0"
                    class="ml-auto w-4 h-4 rounded-full bg-primary text-primary-foreground text-xs font-heading font-semibold flex items-center justify-center flex-shrink-0"
                  >
                    {{ (unreadCounts[ch.id] ?? 0) > 9 ? '9+' : unreadCounts[ch.id] }}
                  </span>
                </NuxtLink>
              </li>

              <li v-if="dmAndGroupChannels.length === 0">
                <button
                  class="flex items-center gap-2 w-full px-4 py-1.5 text-sm font-heading text-muted-foreground/60 hover:text-muted-foreground transition-colors"
                  @click="showNewDM = true"
                >
                  <span class="italic">Nova mensagem directa...</span>
                </button>
              </li>
            </ul>
          </div>
        </nav>

        <!-- Profile section -->
        <div class="flex-shrink-0 border-t border-border px-4 py-3">
          <button
            class="flex items-center gap-2 w-full min-w-0 hover:bg-secondary/50 rounded-md p-1.5 -mx-1.5 transition-colors group"
            @click="showProfile = true"
          >
            <UiAvatar :name="displayName || auth.user_id || 'U'" size="sm" />
            <span class="flex-1 text-sm font-heading text-muted-foreground group-hover:text-foreground truncate text-left">
              {{ displayName || 'Perfil' }}
            </span>
            <svg class="h-3.5 w-3.5 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
          </button>

          <button
            class="mt-1 flex items-center gap-1.5 w-full px-1.5 py-1 text-xs font-heading text-muted-foreground hover:text-foreground transition-colors rounded-md hover:bg-secondary/50"
            @click="handleLogout"
          >
            <svg class="h-3.5 w-3.5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
            </svg>
            Sair
          </button>
        </div>
      </aside>
    </div>

    <!-- Main content -->
    <main class="flex-1 flex flex-col overflow-hidden bg-background">
      <slot />
    </main>

    <!-- Profile Panel overlay -->
    <ProfilePanel
      :open="showProfile"
      @close="onProfileClose"
    />

    <!-- Create Channel Modal -->
    <CreateChannelModal
      :open="showCreateChannel"
      @close="showCreateChannel = false"
    />

    <!-- New DM Modal -->
    <NewDMModal
      :open="showNewDM"
      @close="showNewDM = false"
    />

    <!-- Manage Channel Modal -->
    <ManageChannelModal
      v-if="managingChannelId"
      :open="!!managingChannelId"
      :channel-id="managingChannelId"
      @close="managingChannelId = null"
      @deleted="() => { managingChannelId = null; navigateTo('/') }"
    />

    <!-- Message toasts -->
    <MessageToast :toasts="toasts" @dismiss="dismissToast" />
  </div>
</template>
