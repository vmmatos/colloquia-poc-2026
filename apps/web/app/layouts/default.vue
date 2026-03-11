<script setup lang="ts">
import type { Toast } from '~/components/MessageToast.vue'

const { auth, logout, getProfile } = useAuth()
useTokenRefresh()
const isMobile = useIsMobile()
const { addNotification } = useNotifications()

const displayName = ref('')
const showProfile = ref(false)
const activeChannel = ref('general')
const sidebarOpen = ref(false)

const channels = reactive([
  { id: 'general', label: 'geral', unread: 3 },
  { id: 'random', label: 'random', unread: 1 },
  { id: 'dev', label: 'dev', unread: 0 },
])

const dms = reactive([
  { id: 'alice', label: 'Alice', online: true, unread: 2 },
  { id: 'bob', label: 'Bob', online: true, unread: 0 },
  { id: 'charlie', label: 'Charlie', online: false, unread: 0 },
])

const channelsOpen = ref(true)
const dmsOpen = ref(true)

onMounted(async () => {
  try {
    const profile = await getProfile()
    displayName.value = profile.name || ''
  } catch {
    // profile might not exist yet
  }
})

async function handleLogout() {
  try { await logout() } catch { /* ignore */ }
  await navigateTo('/login')
}

function selectChannel(id: string) {
  const ch = channels.find(c => c.id === id) || dms.find(d => d.id === id)
  if (ch) ch.unread = 0
  activeChannel.value = id
  if (isMobile.value) sidebarOpen.value = false
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

// ---- Simulated incoming messages ----
const SIMULATED = [
  { author: 'Ana Costa',  body: 'Alguém reviu o PR #42?',          channelId: 'dev',    type: 'message' as const },
  { author: 'Bob',        body: '@you o que achas desta proposta?', channelId: 'random', type: 'mention' as const },
  { author: 'Charlie',    body: 'Deploy feito com sucesso 🚀',      channelId: 'dev',    type: 'message' as const },
  { author: 'Ana Costa',  body: 'Reunião às 15h no #geral',         channelId: 'random', type: 'message' as const },
]

const toasts = ref<Toast[]>([])
let simIdx = 0

onMounted(() => {
  const timer = setInterval(() => {
    const sim = SIMULATED[simIdx % SIMULATED.length]
    simIdx++

    if (sim.channelId === activeChannel.value) return

    // Increment unread on channel
    const ch = channels.find(c => c.id === sim.channelId) || dms.find(d => d.id === sim.channelId)
    if (ch) ch.unread++

    // Add to NotificationCenter
    addNotification({
      type: sim.type,
      title: `${sim.author} em #${sim.channelId}`,
      body: sim.body,
      time: 'agora',
    })

    // Show toast (auto-dismiss 3.5s)
    const id = Date.now()
    toasts.value.push({ id, author: sim.author, preview: sim.body.slice(0, 60), channel: `#${sim.channelId}` })
    setTimeout(() => { toasts.value = toasts.value.filter(t => t.id !== id) }, 3500)
  }, 8000)

  onUnmounted(() => clearInterval(timer))
})

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
            <button
              class="flex items-center gap-1 w-full px-4 py-1 text-xs uppercase tracking-wider text-muted-foreground hover:text-foreground transition-colors font-heading"
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

            <ul v-if="channelsOpen" class="mt-1">
              <li v-for="ch in channels" :key="ch.id">
                <NuxtLink
                  :to="ch.id === 'general' ? '/' : `/channels/${ch.id}`"
                  class="flex items-center gap-2 px-4 py-1.5 text-sm font-heading transition-colors"
                  :class="[
                    activeChannel === ch.id
                      ? 'bg-secondary text-foreground'
                      : 'text-muted-foreground hover:bg-secondary/50 hover:text-foreground',
                    ch.unread > 0 && activeChannel !== ch.id ? 'font-semibold text-foreground' : '',
                  ]"
                  @click="selectChannel(ch.id)"
                >
                  <span class="text-muted-foreground">#</span>
                  <span class="flex-1 truncate">{{ ch.label }}</span>
                  <span
                    v-if="ch.unread > 0 && activeChannel !== ch.id"
                    class="ml-auto w-4 h-4 rounded-full bg-primary text-primary-foreground text-xs font-heading font-semibold flex items-center justify-center flex-shrink-0"
                  >
                    {{ ch.unread > 9 ? '9+' : ch.unread }}
                  </span>
                </NuxtLink>
              </li>
            </ul>
          </div>

          <!-- DMs section -->
          <div>
            <button
              class="flex items-center gap-1 w-full px-4 py-1 text-xs uppercase tracking-wider text-muted-foreground hover:text-foreground transition-colors font-heading"
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

            <ul v-if="dmsOpen" class="mt-1">
              <li v-for="dm in dms" :key="dm.id">
                <button
                  class="flex items-center gap-2 w-full px-4 py-1.5 text-sm font-heading transition-colors hover:bg-secondary/50 hover:text-foreground"
                  :class="dm.unread > 0 ? 'text-foreground font-semibold' : 'text-muted-foreground'"
                  @click="selectChannel(dm.id)"
                >
                  <div class="relative flex-shrink-0">
                    <UiAvatar :name="dm.label" size="sm" />
                    <span
                      :class="['absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border border-sidebar',
                        dm.online ? 'bg-emerald-500' : 'bg-muted-foreground/40']"
                    />
                  </div>
                  <span class="flex-1 truncate text-left">{{ dm.label }}</span>
                  <span
                    v-if="dm.unread > 0"
                    class="ml-auto w-4 h-4 rounded-full bg-primary text-primary-foreground text-xs font-heading font-semibold flex items-center justify-center flex-shrink-0"
                  >
                    {{ dm.unread > 9 ? '9+' : dm.unread }}
                  </span>
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
            <!-- Settings icon -->
            <svg class="h-3.5 w-3.5 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
          </button>

          <button
            class="mt-1 flex items-center gap-1.5 w-full px-1.5 py-1 text-xs font-heading text-muted-foreground hover:text-foreground transition-colors rounded-md hover:bg-secondary/50"
            @click="handleLogout"
          >
            <!-- Logout icon -->
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

    <!-- Message toasts -->
    <MessageToast :toasts="toasts" @dismiss="dismissToast" />
  </div>
</template>
