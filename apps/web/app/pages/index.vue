<script setup lang="ts">
definePageMeta({ middleware: 'auth' })

const { auth } = useAuth()
const openSidebar = inject<() => void>('openSidebar')

interface Message {
  id: number
  userId: string
  author: string
  text: string
  time: string
  isAgent?: boolean
}

const messages = ref<Message[]>([
  { id: 1, userId: 'alice', author: 'Alice', text: 'Olá a todos! 👋', time: '10:00' },
  { id: 2, userId: 'bob', author: 'Bob', text: 'Olá Alice! Como estás?', time: '10:01' },
  { id: 3, userId: 'alice', author: 'Alice', text: 'Óptimo, obrigada!', time: '10:02' },
])

const input = ref('')
const isAgentMode = computed(() => input.value.startsWith('@llm'))
const messagesEl = ref<HTMLElement | null>(null)

function scrollToBottom() {
  nextTick(() => {
    if (messagesEl.value) messagesEl.value.scrollTop = messagesEl.value.scrollHeight
  })
}

watch(messages, scrollToBottom, { flush: 'post' })
onMounted(scrollToBottom)

function sendMessage() {
  const text = input.value.trim()
  if (!text) return

  messages.value.push({
    id: Date.now(),
    userId: auth.value.user_id || 'me',
    author: 'Tu',
    text,
    time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
  })
  input.value = ''

  // Mock agent reply if @llm mode
  if (text.startsWith('@llm')) {
    setTimeout(() => {
      messages.value.push({
        id: Date.now() + 1,
        userId: 'llm',
        author: 'LLM',
        text: 'Analisando o contexto da conversa. Esta é uma resposta simulada do agente de IA — a integração real estará disponível em breve.',
        time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
        isAgent: true,
      })
    }, 800)
  }
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    sendMessage()
  }
}
</script>

<template>
  <div class="flex flex-col h-full bg-background">
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
        <span class="font-heading font-semibold text-foreground text-sm">geral</span>
      </div>
      <NotificationCenter />
    </header>

    <!-- Messages -->
    <div ref="messagesEl" class="flex-1 overflow-y-auto px-4 md:px-6 py-4 space-y-1">
      <div
        v-for="msg in messages"
        :key="msg.id"
        :class="['animate-fade-in py-2 group', msg.isAgent ? 'pl-4 border-l-2 border-primary' : '']"
      >
        <div class="flex items-baseline gap-2">
          <span
            :class="['font-heading text-sm font-semibold', msg.isAgent ? 'text-primary' : 'text-foreground']"
          >
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

      <p v-if="messages.length === 0" class="text-muted-foreground font-body text-sm italic text-center py-12">
        Sem mensagens ainda. Começa a conversa!
      </p>
    </div>

    <!-- Message input -->
    <div class="px-4 md:px-6 pb-4 pt-2 flex-shrink-0">
      <div
        :class="[
          'flex items-end gap-2 rounded-lg border bg-surface px-4 py-3 transition-all',
          isAgentMode
            ? 'border-primary shadow-[0_0_12px_-4px_hsl(43_100%_50%_/_0.3)]'
            : 'border-border'
        ]"
      >
        <textarea
          v-model="input"
          rows="1"
          :placeholder="isAgentMode ? 'Pergunta ao agente LLM...' : 'Mensagem para #geral'"
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
          :disabled="!input.trim()"
          :class="['text-muted-foreground transition-colors flex-shrink-0 pb-0.5', input.trim() ? 'hover:text-primary' : 'opacity-30']"
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
</template>
