<script setup lang="ts">
import type { UserProfile } from '../../shared/types/auth'

defineProps<{ open: boolean }>()
const emit = defineEmits<{ close: [] }>()

const { t } = useI18n()
const { createDM } = useChannels()
const { auth } = useAuth()
const config = useRuntimeConfig()

const query = ref('')
const results = ref<UserProfile[]>([])
const searchLoading = ref(false)
const loading = ref(false)
const error = ref<string | null>(null)
let searchTimer: ReturnType<typeof setTimeout> | null = null

watch(query, (q) => {
  if (searchTimer) clearTimeout(searchTimer)
  results.value = []
  if (q.trim().length < 2) return
  searchTimer = setTimeout(async () => {
    searchLoading.value = true
    try {
      const res = await $fetch<UserProfile[]>(`${config.public.apiBase}/api/v1/users/search`, {
        query: { q: q.trim() },
        headers: { Authorization: `Bearer ${auth.value.access_token}` },
      })
      // Filter out the current user
      results.value = (res ?? []).filter(u => u.user_id !== auth.value.user_id)
    } catch {
      results.value = []
    } finally {
      searchLoading.value = false
    }
  }, 300)
})

async function selectUser(user: UserProfile) {
  error.value = null
  loading.value = true
  try {
    const ch = await createDM(user.user_id)
    close()
    await navigateTo(`/channels/${ch.id}`)
  } catch (err: unknown) {
    const msg = (err as { data?: { error?: string } })?.data?.error
    error.value = msg ?? t('dm.error')
  } finally {
    loading.value = false
  }
}

function close() {
  query.value = ''
  results.value = []
  error.value = null
  emit('close')
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
        <div class="relative z-10 w-full max-w-sm bg-card border border-border rounded-lg shadow-xl">
          <!-- Header -->
          <div class="flex items-center justify-between px-6 py-4 border-b border-border">
            <h2 class="font-heading font-semibold text-foreground text-base">{{ $t('dm.title') }}</h2>
            <button
              class="text-muted-foreground hover:text-foreground transition-colors"
              @click="close"
            >
              <svg class="h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- Body -->
          <div class="px-6 py-5 space-y-3">
            <div class="relative">
              <input
                v-model="query"
                type="text"
                :placeholder="$t('dm.searchPlaceholder')"
                autofocus
                class="w-full bg-background border border-border rounded-md px-3 py-2 text-sm font-heading text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary/50 transition-colors"
              />
              <!-- Search spinner -->
              <span
                v-if="searchLoading"
                class="absolute right-3 top-1/2 -translate-y-1/2 flex gap-0.5"
              >
                <span class="w-1 h-1 rounded-full bg-muted-foreground animate-bounce [animation-delay:-0.3s]" />
                <span class="w-1 h-1 rounded-full bg-muted-foreground animate-bounce [animation-delay:-0.15s]" />
                <span class="w-1 h-1 rounded-full bg-muted-foreground animate-bounce" />
              </span>
            </div>

            <!-- Results -->
            <ul v-if="results.length > 0" class="space-y-1 max-h-48 overflow-y-auto">
              <li
                v-for="u in results"
                :key="u.user_id"
                class="flex items-center gap-3 px-3 py-2 rounded-md cursor-pointer hover:bg-secondary/50 transition-colors"
                :class="{ 'opacity-50 pointer-events-none': loading }"
                @click="selectUser(u)"
              >
                <UiAvatar :name="u.name || u.user_id" size="sm" />
                <span class="text-sm font-heading text-foreground">{{ u.name || u.user_id }}</span>
              </li>
            </ul>

            <p
              v-else-if="query.trim().length >= 2 && !searchLoading"
              class="text-xs font-body text-muted-foreground text-center py-2"
            >
              {{ $t('dm.noResults') }}
            </p>

            <p v-if="error" class="text-xs text-destructive font-body">{{ error }}</p>
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
