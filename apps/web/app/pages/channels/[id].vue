<script setup lang="ts">
import type { Channel, ChannelMember } from '../../../shared/types/channels'
import type { Message } from '~/composables/useMessaging'
import type { SseEvent } from '~/composables/useSSE'

definePageMeta({ middleware: 'auth' })

const { auth } = useAuth()
const { fetchChannel, fetchMembers } = useChannels()
const { fetchMessages, sendMessage: apiSendMessage } = useMessaging()
const { resolveUser, prefetchUsers } = useUsersCache()
const { getPeer } = useDMPeers()
const { suggestions, isLoading: isSuggestionsLoading, debouncedFetch, clearSuggestions } = useAssist()
const openSidebar = inject<() => void>('openSidebar')
const registerChannelHandler = inject<(fn: ((e: SseEvent) => void) | null) => void>('registerChannelHandler')
const route = useRoute()

const channelId = computed(() => route.params.id as string)

const channel = ref<Channel | null>(null)
const myRole = ref<string>('')
const memberIds = ref<string[]>([])
const showManage = ref(false)
const showNewGroup = ref(false)
const loadError = ref('')
const isSending = ref(false)

interface DisplayMessage {
  id: string
  userId: string
  author: string
  text: string
  time: string
  isAgent?: boolean
}

const messages = ref<DisplayMessage[]>([])
const input = ref('')
const isAgentMode = computed(() => input.value.startsWith('@llm'))
const messagesEl = ref<HTMLElement | null>(null)
const textareaRef = ref<HTMLTextAreaElement | null>(null)

// ── Channel type helpers ───────────────────────────────────────────────────────

const isDM = computed(() => channel.value?.type === 'dm')
const isGroup = computed(() => channel.value?.type === 'group')

const peerUserId = computed(() => isDM.value ? getPeer(channelId.value) : null)
const peerName = computed(() => peerUserId.value ? resolveUser(peerUserId.value) : '...')

const channelDisplayName = computed(() => {
  if (isDM.value) return peerName.value
  if (isGroup.value) return channel.value?.name || 'Grupo'
  return channel.value?.name ?? '...'
})

const inputPlaceholder = computed(() => {
  if (isAgentMode.value) return 'Pergunta ao agente LLM...'
  if (isDM.value) return `Mensagem para ${peerName.value}`
  if (isGroup.value) return `Mensagem para ${channel.value?.name || 'o grupo'}`
  return `Mensagem para #${channel.value?.name ?? ''}`
})

// ── Scroll ─────────────────────────────────────────────────────────────────────

function scrollToBottom() {
  nextTick(() => {
    if (messagesEl.value) messagesEl.value.scrollTop = messagesEl.value.scrollHeight
  })
}

watch(() => messages.value.length, scrollToBottom, { flush: 'post' })

function formatTime(unixSeconds: number): string {
  return new Date(unixSeconds * 1000).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function toDisplay(m: Message): DisplayMessage {
  return {
    id: m.id,
    userId: m.user_id,
    author: resolveUser(m.user_id),
    text: m.content,
    time: formatTime(m.created_at),
  }
}

// ── Load channel ───────────────────────────────────────────────────────────────

async function loadChannel() {
  loadError.value = ''
  channel.value = null
  myRole.value = ''
  memberIds.value = []
  messages.value = []

  try {
    const [ch, memberList, history] = await Promise.all([
      fetchChannel(channelId.value),
      fetchMembers(channelId.value),
      fetchMessages(channelId.value),
    ])
    channel.value = ch
    const members = memberList as ChannelMember[]
    const me = members.find(m => m.user_id === auth.value.user_id)
    myRole.value = me?.role ?? ''
    memberIds.value = members.map(m => m.user_id)
    await prefetchUsers(memberIds.value)
    messages.value = history.map(toDisplay)
  } catch {
    loadError.value = 'Canal não encontrado ou sem acesso.'
  }
}

onMounted(() => {
  loadChannel()
  window.visualViewport?.addEventListener('resize', scrollToBottom)
  registerChannelHandler?.((event) => {
    if (messages.value.some(m => m.id === event.id)) return // dedup com optimistic
    messages.value.push({
      id: event.id,
      userId: event.user_id,
      author: resolveUser(event.user_id),
      text: event.content,
      time: formatTime(event.created_at),
    })
  })
})

onUnmounted(() => {
  registerChannelHandler?.(null)
  window.visualViewport?.removeEventListener('resize', scrollToBottom)
})

watch(channelId, loadChannel)
watch(input, val => debouncedFetch(channelId.value, val))
watch(() => suggestions.value.length, (newLen, oldLen) => {
  if (newLen > 0 && oldLen === 0) scrollToBottom()
})

// ── Send message ───────────────────────────────────────────────────────────────

async function sendMessage() {
  const text = input.value.trim()
  if (!text || isSending.value) return

  // Mock do agente LLM mantém-se local
  if (text.startsWith('@llm')) {
    messages.value.push({
      id: `local-${Date.now()}`,
      userId: auth.value.user_id || 'me',
      author: 'Tu',
      text,
      time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
    })
    input.value = ''
    setTimeout(() => {
      messages.value.push({
        id: `local-${Date.now() + 1}`,
        userId: 'llm',
        author: 'LLM',
        text: 'Analisando o contexto da conversa. Esta é uma resposta simulada do agente de IA — a integração real estará disponível em breve.',
        time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
        isAgent: true,
      })
    }, 800)
    return
  }

  isSending.value = true
  try {
    const sent = await apiSendMessage(channelId.value, text)
    input.value = ''
    clearSuggestions()
    messages.value.push(toDisplay(sent))
  } catch {
    // Falha silenciosa no POC — input não é limpo
  } finally {
    isSending.value = false
  }
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    clearSuggestions()
    return
  }
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    sendMessage()
  }
}

function applySuggestion(suggestion: string) {
  clearSuggestions()
  input.value = suggestion
  nextTick(() => {
    if (textareaRef.value) {
      textareaRef.value.focus()
      textareaRef.value.style.height = 'auto'
      textareaRef.value.style.height = textareaRef.value.scrollHeight + 'px'
    }
  })
}

const isAdminOrOwner = computed(() => myRole.value === 'owner' || myRole.value === 'admin')
</script>

<template>
  <div class="flex flex-col h-full bg-background">
    <!-- Error state -->
    <div v-if="loadError" class="flex-1 flex items-center justify-center">
      <p class="text-muted-foreground font-body text-sm">{{ loadError }}</p>
    </div>

    <div v-else class="flex flex-col flex-1 min-h-0">
      <!-- Channel header -->
      <header class="h-14 flex items-center justify-between px-4 md:px-6 border-b border-border flex-shrink-0">
        <div class="flex items-center gap-2">
          <!-- Mobile menu button -->
          <button
            class="md:hidden text-muted-foreground hover:text-foreground transition-colors mr-1"
            @click="openSidebar?.()"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <line x1="4" x2="20" y1="6" y2="6" />
              <line x1="4" x2="20" y1="12" y2="12" />
              <line x1="4" x2="20" y1="18" y2="18" />
            </svg>
          </button>

          <!-- DM header: avatar + peer name -->
          <template v-if="isDM">
            <UiAvatar :name="peerName" size="sm" />
            <span class="font-heading font-semibold text-foreground text-sm">{{ peerName }}</span>
          </template>

          <!-- Group header: people icon + name -->
          <template v-else-if="isGroup">
            <svg class="h-4 w-4 text-muted-foreground flex-shrink-0" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            <span class="font-heading font-semibold text-foreground text-sm">{{ channelDisplayName }}</span>
          </template>

          <!-- Regular channel header: # + name -->
          <template v-else>
            <span class="text-muted-foreground">#</span>
            <span class="font-heading font-semibold text-foreground text-sm">{{ channel?.name ?? '...' }}</span>
          </template>
        </div>

        <div class="flex items-center gap-2">
          <!-- Add people button (DM only) -->
          <button
            v-if="isDM && channel"
            class="text-muted-foreground hover:text-foreground transition-colors"
            title="Adicionar pessoas à conversa"
            @click="showNewGroup = true"
          >
            <svg class="h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
            </svg>
          </button>

          <!-- Manage channel button (admin/owner, non-DM) -->
          <button
            v-if="isAdminOrOwner && channel && !isDM"
            class="text-muted-foreground hover:text-foreground transition-colors"
            title="Gerir canal"
            @click="showManage = true"
          >
            <svg class="h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
          </button>
          <NotificationCenter />
        </div>
      </header>

      <!-- Messages -->
      <div ref="messagesEl" class="flex-1 overflow-y-auto px-4 md:px-6 py-4 space-y-1">
        <div
          v-for="msg in messages"
          :key="msg.id"
          :class="['animate-fade-in py-2 group', msg.isAgent ? 'pl-4 border-l-2 border-primary' : '']"
        >
          <div class="flex items-baseline gap-2">
            <span :class="['font-heading text-sm font-semibold', msg.isAgent ? 'text-primary' : 'text-foreground']">
              {{ msg.author }}
            </span>
            <span class="text-muted-foreground text-xs opacity-0 group-hover:opacity-100 transition-opacity">
              {{ msg.time }}
            </span>
          </div>
          <p :class="['text-sm leading-relaxed mt-0.5', msg.isAgent ? 'font-body italic text-foreground/90' : 'font-body text-foreground']">
            {{ msg.text }}
          </p>
        </div>

        <div v-if="messages.length === 0 && channel" class="min-h-full flex items-center justify-center">
          <p class="text-muted-foreground font-body text-sm italic text-center">
            Sem mensagens ainda. Começa a conversa!
          </p>
        </div>
      </div>

      <!-- Archived channel banner -->
      <div v-if="channel?.archived" class="px-4 md:px-6 py-2 text-xs text-muted-foreground bg-surface border-t border-border font-body italic">
        Este canal está arquivado. Não é possível enviar mensagens.
      </div>

      <!-- Message input -->
      <div v-if="!channel?.archived" class="px-4 md:px-6 pb-4 pt-2 flex-shrink-0">
        <!-- Suggestion pills strip -->
        <Transition
          enter-active-class="transition-all duration-200 ease-out"
          enter-from-class="opacity-0 translate-y-1"
          leave-active-class="transition-all duration-150 ease-in"
          leave-to-class="opacity-0 translate-y-1"
        >
          <div
            v-if="suggestions.length > 0"
            class="flex items-center gap-2 overflow-x-auto pb-2"
            style="scrollbar-width: none; -ms-overflow-style: none;"
          >
            <button
              v-for="(suggestion, i) in suggestions"
              :key="i"
              class="flex-shrink-0 px-3 py-1 rounded-full border border-border bg-surface text-sm font-heading text-muted-foreground hover:border-primary hover:text-primary hover:bg-primary/5 transition-colors"
              @mousedown.prevent
              @click="applySuggestion(suggestion)"
            >
              {{ suggestion }}
            </button>
          </div>
        </Transition>

        <div
          :class="[
            'flex items-end gap-2 rounded-lg border bg-surface px-4 py-3 transition-all',
            isAgentMode
              ? 'border-primary shadow-[0_0_12px_-4px_hsl(43_100%_50%_/_0.3)]'
              : 'border-border',
          ]"
        >
          <textarea
            ref="textareaRef"
            v-model="input"
            rows="1"
            :placeholder="inputPlaceholder"
            :class="[
              'flex-1 bg-transparent resize-none outline-none text-sm text-foreground',
              'placeholder:text-muted-foreground leading-relaxed',
              'max-h-32 overflow-y-auto',
              isAgentMode ? 'font-body italic' : 'font-heading',
            ]"
            @keydown="handleKeydown"
            @focus="scrollToBottom"
            @blur="clearSuggestions"
            @input="($event.target as HTMLTextAreaElement).style.height = 'auto'; ($event.target as HTMLTextAreaElement).style.height = ($event.target as HTMLTextAreaElement).scrollHeight + 'px'"
          />

          <!-- Loading dots -->
          <span
            v-if="isSuggestionsLoading"
            class="flex items-end gap-0.5 pb-0.5 flex-shrink-0"
            aria-label="A carregar sugestões"
          >
            <span class="w-1 h-1 rounded-full bg-muted-foreground animate-bounce [animation-delay:-0.3s]" />
            <span class="w-1 h-1 rounded-full bg-muted-foreground animate-bounce [animation-delay:-0.15s]" />
            <span class="w-1 h-1 rounded-full bg-muted-foreground animate-bounce" />
          </span>

          <button
            :disabled="!input.trim() || isSending"
            :class="['text-muted-foreground transition-colors flex-shrink-0 pb-0.5', input.trim() && !isSending ? 'hover:text-primary' : 'opacity-30']"
            @click="sendMessage"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="m22 2-7 20-4-9-9-4 20-7"/>
              <path d="M22 2 11 13"/>
            </svg>
          </button>
        </div>
        <p v-if="isAgentMode" class="text-xs text-primary/70 font-heading mt-1.5 ml-1">
          Modo agente activo — a resposta será gerada por LLM
        </p>
      </div>
    </div>

    <!-- Manage Channel Modal (non-DM channels) -->
    <ManageChannelModal
      v-if="channel && !isDM"
      :open="showManage"
      :channel-id="channelId"
      @close="showManage = false"
      @deleted="navigateTo('/')"
    />

    <!-- New Group Modal (from DM) -->
    <NewGroupModal
      v-if="isDM"
      :open="showNewGroup"
      :existing-member-ids="memberIds"
      @close="showNewGroup = false"
    />
  </div>
</template>
