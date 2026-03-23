<script setup lang="ts">
import type { CreateChannelInput } from '../../shared/types/channels'
import type { UserProfile } from '../../shared/types/auth'

defineProps<{ open: boolean }>()
const emit = defineEmits<{ close: [] }>()

const { createChannel } = useChannels()
const { auth } = useAuth()
const config = useRuntimeConfig()

const name = ref('')
const description = ref('')
const isPrivate = ref(false)
const loading = ref(false)
const error = ref<string | null>(null)

// Member search
const memberQuery = ref('')
const memberResults = ref<UserProfile[]>([])
const selectedMembers = ref<UserProfile[]>([])
const searchLoading = ref(false)
let searchTimer: ReturnType<typeof setTimeout> | null = null

watch(memberQuery, (q) => {
  if (searchTimer) clearTimeout(searchTimer)
  if (!q.trim()) { memberResults.value = []; return }
  searchTimer = setTimeout(async () => {
    searchLoading.value = true
    try {
      const res = await $fetch<UserProfile[]>(`${config.public.apiBase}/api/v1/users/search`, {
        query: { q: q.trim() },
        headers: { Authorization: `Bearer ${auth.value.access_token}` },
      })
      memberResults.value = res ?? []
    } catch {
      memberResults.value = []
    } finally {
      searchLoading.value = false
    }
  }, 300)
})

function selectMember(user: UserProfile) {
  if (!selectedMembers.value.find(m => m.user_id === user.user_id)) {
    selectedMembers.value = [...selectedMembers.value, user]
  }
  memberQuery.value = ''
  memberResults.value = []
}

function removeMember(userId: string) {
  selectedMembers.value = selectedMembers.value.filter(m => m.user_id !== userId)
}

function reset() {
  name.value = ''
  description.value = ''
  isPrivate.value = false
  error.value = null
  memberQuery.value = ''
  memberResults.value = []
  selectedMembers.value = []
}

function close() {
  reset()
  emit('close')
}

async function submit() {
  error.value = null
  const trimmedName = name.value.trim()
  if (!trimmedName) {
    error.value = 'O nome do canal é obrigatório.'
    return
  }

  loading.value = true
  try {
    const input: CreateChannelInput = {
      name: trimmedName,
      description: description.value.trim() || undefined,
      is_private: isPrivate.value,
      type: 'channel',
      ...(selectedMembers.value.length > 0 && {
        initial_member_ids: selectedMembers.value.map(m => m.user_id),
      }),
    }
    const ch = await createChannel(input)
    close()
    await navigateTo(`/channels/${ch.id}`)
  } catch (err: unknown) {
    const msg = (err as { data?: { error?: string } })?.data?.error
    error.value = msg ?? 'Erro ao criar o canal. Tenta novamente.'
  } finally {
    loading.value = false
  }
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
        <div class="relative z-10 w-full max-w-md bg-card border border-border rounded-lg shadow-xl">
          <!-- Header -->
          <div class="flex items-center justify-between px-6 py-4 border-b border-border">
            <h2 class="font-heading font-semibold text-foreground text-base">Criar canal</h2>
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
          <div class="px-6 py-5 space-y-4">
            <!-- Name -->
            <div>
              <label class="block text-xs font-heading font-medium text-muted-foreground uppercase tracking-wider mb-1.5">
                Nome <span class="text-destructive">*</span>
              </label>
              <input
                v-model="name"
                type="text"
                placeholder="ex: marketing, anúncios"
                maxlength="80"
                class="w-full bg-background border border-border rounded-md px-3 py-2 text-sm font-heading text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary/50 transition-colors"
                @keydown.enter="submit"
              />
            </div>

            <!-- Description -->
            <div>
              <label class="block text-xs font-heading font-medium text-muted-foreground uppercase tracking-wider mb-1.5">
                Descrição <span class="text-muted-foreground font-normal">(opcional)</span>
              </label>
              <input
                v-model="description"
                type="text"
                placeholder="Para que serve este canal?"
                maxlength="255"
                class="w-full bg-background border border-border rounded-md px-3 py-2 text-sm font-heading text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary/50 transition-colors"
              />
            </div>

            <!-- Private toggle -->
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm font-heading text-foreground">Canal privado</p>
                <p class="text-xs font-body text-muted-foreground mt-0.5">Só quem for adicionado pode ver e entrar</p>
              </div>
              <button
                :class="[
                  'relative inline-flex h-6 w-11 flex-shrink-0 items-center rounded-full transition-colors',
                  isPrivate ? 'bg-primary' : 'bg-secondary',
                ]"
                @click="isPrivate = !isPrivate"
              >
                <span
                  :class="[
                    'inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform',
                    isPrivate ? 'translate-x-6' : 'translate-x-1',
                  ]"
                />
              </button>
            </div>

            <!-- Initial members -->
            <div>
              <label class="block text-xs font-heading font-medium text-muted-foreground uppercase tracking-wider mb-1.5">
                Membros iniciais <span class="text-muted-foreground font-normal">(opcional)</span>
              </label>
              <!-- Selected members chips -->
              <div v-if="selectedMembers.length > 0" class="flex flex-wrap gap-1.5 mb-2">
                <span
                  v-for="m in selectedMembers"
                  :key="m.user_id"
                  class="inline-flex items-center gap-1 px-2 py-0.5 bg-secondary rounded-full text-xs font-heading text-foreground"
                >
                  {{ m.name || m.user_id }}
                  <button
                    type="button"
                    class="text-muted-foreground hover:text-foreground ml-0.5"
                    @click="removeMember(m.user_id)"
                  >
                    <svg class="h-3 w-3" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </span>
              </div>
              <!-- Search input -->
              <div class="relative">
                <input
                  v-model="memberQuery"
                  type="text"
                  placeholder="Pesquisar utilizadores..."
                  class="w-full bg-background border border-border rounded-md px-3 py-2 text-sm font-heading text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary/50 transition-colors"
                />
                <ul
                  v-if="memberResults.length > 0"
                  class="absolute z-10 mt-1 w-full bg-card border border-border rounded-md shadow-lg max-h-40 overflow-y-auto"
                >
                  <li
                    v-for="u in memberResults"
                    :key="u.user_id"
                    class="flex items-center gap-2 px-3 py-2 cursor-pointer hover:bg-secondary/50 text-sm font-heading text-foreground"
                    @mousedown.prevent="selectMember(u)"
                  >
                    <UiAvatar :name="u.name || u.user_id" size="sm" />
                    <span>{{ u.name || u.user_id }}</span>
                  </li>
                </ul>
              </div>
            </div>

            <!-- Error -->
            <p v-if="error" class="text-xs text-destructive font-body">{{ error }}</p>
          </div>

          <!-- Footer -->
          <div class="flex justify-end gap-3 px-6 py-4 border-t border-border">
            <button
              class="px-4 py-2 text-sm font-heading text-muted-foreground hover:text-foreground transition-colors"
              @click="close"
            >
              Cancelar
            </button>
            <UiButton :loading="loading" @click="submit">
              Criar canal
            </UiButton>
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
