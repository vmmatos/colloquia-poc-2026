<script setup lang="ts">
import type { Channel, ChannelMember } from '../../../shared/types/channels'
import type { Message } from '~/composables/useMessaging'
import type { SseEvent } from '~/composables/useSSE'

definePageMeta({ middleware: 'auth' })

const { auth } = useAuth()
const { fetchChannel, fetchMembers } = useChannels()
const { fetchMessages, sendMessage: apiSendMessage } = useMessaging()
const { resolveUser, prefetchUsers } = useUsersCache()
const openSidebar = inject<() => void>('openSidebar')
const registerChannelHandler = inject<(fn: ((e: SseEvent) => void) | null) => void>('registerChannelHandler')
const route = useRoute()

const channelId = computed(() => route.params.id as string)

const channel = ref<Channel | null>(null)
const myRole = ref<string>('')
const showManage = ref(false)
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

async function loadChannel() {
  loadError.value = ''
  channel.value = null
  myRole.value = ''
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
    await prefetchUsers(members.map(m => m.user_id))
    messages.value = history.map(toDisplay)
  } catch {
    loadError.value = 'Canal não encontrado ou sem acesso.'
  }
}

onMounted(() => {
  loadChannel()
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
})

watch(channelId, loadChannel)

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
    await apiSendMessage(channelId.value, text)
    input.value = ''
  } catch {
    // Falha silenciosa no POC — input não é limpo
  } finally {
    isSending.value = false
  }
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    sendMessage()
  }
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
          <span class="text-muted-foreground">#</span>
          <span class="font-heading font-semibold text-foreground text-sm">{{ channel?.name ?? '...' }}</span>
        </div>
        <div class="flex items-center gap-2">
          <button
            v-if="isAdminOrOwner && channel"
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
        <div
          :class="[
            'flex items-end gap-2 rounded-lg border bg-surface px-4 py-3 transition-all',
            isAgentMode
              ? 'border-primary shadow-[0_0_12px_-4px_hsl(43_100%_50%_/_0.3)]'
              : 'border-border',
          ]"
        >
          <textarea
            v-model="input"
            rows="1"
            :placeholder="isAgentMode ? 'Pergunta ao agente LLM...' : `Mensagem para #${channel?.name ?? ''}`"
            :class="[
              'flex-1 bg-transparent resize-none outline-none text-sm text-foreground',
              'placeholder:text-muted-foreground leading-relaxed',
              'max-h-32 overflow-y-auto',
              isAgentMode ? 'font-body italic' : 'font-heading',
            ]"
            @keydown="handleKeydown"
            @input="($event.target as HTMLTextAreaElement).style.height = 'auto'; ($event.target as HTMLTextAreaElement).style.height = ($event.target as HTMLTextAreaElement).scrollHeight + 'px'"
          />
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

    <!-- Manage Channel Modal -->
    <ManageChannelModal
      v-if="channel"
      :open="showManage"
      :channel-id="channelId"
      @close="showManage = false"
      @deleted="navigateTo('/')"
    />
  </div>
</template>
