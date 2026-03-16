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

const { auth } = useAuth()
const { fetchMembers, addMember, removeMember, deleteChannel } = useChannels()
const config = useRuntimeConfig()

const channel = ref<Channel | null>(null)
const members = ref<ChannelMember[]>([])
const myRole = ref<string>('')
const activeTab = ref<'members' | 'settings'>('members')
const loadError = ref('')

// Add member form
const lookupId = ref('')
const lookedUpUser = ref<UserProfile | null>(null)
const lookupError = ref('')
const lookupLoading = ref(false)
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
  channel.value = null
  myRole.value = ''
  activeTab.value = 'members'
  lookedUpUser.value = null
  lookupId.value = ''
  lookupError.value = ''
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
  } catch {
    loadError.value = 'Erro ao carregar dados do canal.'
  }
}

watch(() => props.open, (val) => {
  if (val) loadData()
}, { immediate: true })

watch(() => props.channelId, () => {
  if (props.open) loadData()
})

function close() {
  emit('close')
}

async function doLookup() {
  lookupError.value = ''
  lookedUpUser.value = null
  const id = lookupId.value.trim()
  if (!id) return

  lookupLoading.value = true
  try {
    lookedUpUser.value = await $fetch<UserProfile>(`${config.public.apiBase}/api/v1/users/${id}`, {
      headers: { Authorization: `Bearer ${auth.value.access_token}` },
    })
  } catch {
    lookupError.value = 'Utilizador não encontrado.'
  } finally {
    lookupLoading.value = false
  }
}

async function doAddMember() {
  if (!lookedUpUser.value) return
  addError.value = ''
  addLoading.value = true
  try {
    const newMember = await addMember(props.channelId, { user_id: lookedUpUser.value.user_id })
    members.value = [...members.value, newMember]
    lookedUpUser.value = null
    lookupId.value = ''
  } catch (err: unknown) {
    const msg = (err as { data?: { error?: string } })?.data?.error
    addError.value = msg ?? 'Erro ao adicionar membro.'
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
    deleteError.value = msg ?? 'Erro ao eliminar o canal.'
  } finally {
    deleteLoading.value = false
  }
}

function roleBadge(role: string) {
  if (role === 'owner') return { label: 'owner', cls: 'bg-primary/20 text-primary' }
  if (role === 'admin') return { label: 'admin', cls: 'bg-secondary text-foreground' }
  return { label: 'membro', cls: 'bg-secondary/50 text-muted-foreground' }
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
              Membros <span class="text-xs text-muted-foreground ml-1">({{ members.length }})</span>
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
              Configurações
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
                <UiAvatar :name="m.user_id" size="sm" />
                <div class="flex-1 min-w-0">
                  <p class="text-sm font-heading text-foreground truncate">{{ m.user_id }}</p>
                </div>
                <span :class="['text-xs px-2 py-0.5 rounded-full font-heading', roleBadge(m.role).cls]">
                  {{ roleBadge(m.role).label }}
                </span>
                <button
                  v-if="(myRole === 'owner' || myRole === 'admin') && m.user_id !== auth.user_id && m.role !== 'owner'"
                  :disabled="removingUserId === m.user_id"
                  class="text-xs font-heading text-muted-foreground hover:text-destructive transition-colors opacity-0 group-hover:opacity-100 disabled:opacity-30"
                  @click="doRemoveMember(m.user_id)"
                >
                  Remover
                </button>
              </li>
            </ul>

            <!-- Add member section -->
            <div v-if="myRole === 'owner' || myRole === 'admin'" class="border-t border-border px-4 py-4">
              <p class="text-xs font-heading font-medium text-muted-foreground uppercase tracking-wider mb-3">
                Adicionar membro
              </p>
              <div class="flex gap-2">
                <input
                  v-model="lookupId"
                  type="text"
                  placeholder="User ID (UUID)"
                  class="flex-1 bg-background border border-border rounded-md px-3 py-2 text-sm font-heading text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary/50 transition-colors"
                  @keydown.enter="doLookup"
                />
                <UiButton variant="secondary" :loading="lookupLoading" @click="doLookup">
                  Procurar
                </UiButton>
              </div>
              <p v-if="lookupError" class="text-xs text-destructive font-body mt-1.5">{{ lookupError }}</p>

              <!-- Lookup result -->
              <div v-if="lookedUpUser" class="mt-3 flex items-center gap-3 p-3 bg-secondary/30 rounded-md">
                <UiAvatar :name="lookedUpUser.name || lookedUpUser.user_id" size="sm" />
                <div class="flex-1 min-w-0">
                  <p class="text-sm font-heading text-foreground truncate">{{ lookedUpUser.name || 'Sem nome' }}</p>
                  <p class="text-xs font-body text-muted-foreground truncate">{{ lookedUpUser.user_id }}</p>
                </div>
                <UiButton :loading="addLoading" @click="doAddMember">
                  Adicionar
                </UiButton>
              </div>
              <p v-if="addError" class="text-xs text-destructive font-body mt-1.5">{{ addError }}</p>
            </div>
          </div>

          <!-- Settings tab (owner only) -->
          <div v-else-if="activeTab === 'settings'" class="flex-1 overflow-y-auto px-6 py-5">
            <div class="border border-destructive/30 rounded-lg p-5">
              <h3 class="font-heading font-semibold text-foreground text-sm mb-1">Eliminar canal</h3>
              <p class="text-xs font-body text-muted-foreground mb-4">
                Esta ação é irreversível. O canal e todo o seu conteúdo serão permanentemente eliminados.
              </p>

              <div v-if="!confirmDelete">
                <button
                  class="px-4 py-2 text-sm font-heading text-destructive border border-destructive/40 rounded-md hover:bg-destructive/10 transition-colors"
                  @click="confirmDelete = true"
                >
                  Eliminar canal
                </button>
              </div>

              <div v-else class="space-y-3">
                <p class="text-sm font-heading text-foreground">
                  Tens a certeza? Escreve o nome do canal para confirmar:
                </p>
                <p class="font-mono text-sm text-primary">{{ channel?.name }}</p>
                <div class="flex gap-3">
                  <UiButton variant="danger" :loading="deleteLoading" @click="doDelete">
                    Sim, eliminar
                  </UiButton>
                  <button
                    class="px-4 py-2 text-sm font-heading text-muted-foreground hover:text-foreground transition-colors"
                    @click="confirmDelete = false"
                  >
                    Cancelar
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
