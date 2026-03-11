<script setup lang="ts">
const { notifications, unreadCount, markAllRead, markRead } = useNotifications()

const open = ref(false)

// Close on click outside
const containerRef = ref<HTMLElement | null>(null)
onMounted(() => {
  document.addEventListener('click', onOutsideClick)
})
onUnmounted(() => {
  document.removeEventListener('click', onOutsideClick)
})
function onOutsideClick(e: MouseEvent) {
  if (containerRef.value && !containerRef.value.contains(e.target as Node)) {
    open.value = false
  }
}
</script>

<template>
  <div ref="containerRef" class="relative">
    <!-- Bell button -->
    <button
      class="relative p-2 rounded-md text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-colors"
      @click="open = !open"
    >
      <!-- Bell icon -->
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
      </svg>
      <!-- Badge -->
      <span
        v-if="unreadCount > 0"
        class="absolute -top-0.5 -right-0.5 h-4 w-4 rounded-full bg-primary text-primary-foreground text-xs font-heading font-semibold flex items-center justify-center"
      >
        {{ unreadCount > 9 ? '9+' : unreadCount }}
      </span>
    </button>

    <!-- Dropdown -->
    <div
      v-if="open"
      class="absolute right-0 top-full mt-2 w-80 max-h-96 bg-card border border-border rounded-lg shadow-xl z-50 flex flex-col animate-fade-in"
    >
      <!-- Header -->
      <div class="flex items-center justify-between px-4 py-3 border-b border-border flex-shrink-0">
        <span class="text-sm font-heading font-semibold text-foreground">Notificações</span>
        <button
          v-if="unreadCount > 0"
          class="text-xs text-muted-foreground hover:text-primary transition-colors font-heading"
          @click="markAllRead"
        >
          Marcar todas
        </button>
      </div>

      <!-- List -->
      <div class="overflow-y-auto flex-1">
        <div
          v-for="n in notifications"
          :key="n.id"
          class="flex items-start gap-3 px-4 py-3 hover:bg-secondary/50 cursor-pointer transition-colors"
          @click="markRead(n.id)"
        >
          <!-- Icon -->
          <div class="flex-shrink-0 mt-0.5">
            <!-- AtSign for mention -->
            <svg v-if="n.type === 'mention'" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 12a4 4 0 10-8 0 4 4 0 008 0zm0 0v1.5a2.5 2.5 0 005 0V12a9 9 0 10-9 9m4.5-1.206a8.959 8.959 0 01-4.5 1.207" />
            </svg>
            <!-- Bot for agent -->
            <svg v-else-if="n.type === 'agent'" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17H3a2 2 0 01-2-2V5a2 2 0 012-2h14a2 2 0 012 2v10a2 2 0 01-2 2h-2" />
            </svg>
            <!-- Message for generic -->
            <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-muted-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
          </div>

          <!-- Content -->
          <div class="flex-1 min-w-0">
            <p :class="['text-xs font-heading truncate', n.read ? 'text-muted-foreground' : 'text-foreground font-semibold']">
              {{ n.title }}
            </p>
            <p class="text-xs text-muted-foreground font-body truncate mt-0.5">{{ n.body }}</p>
            <p class="text-xs text-muted-foreground/60 font-heading mt-1">{{ n.time }}</p>
          </div>

          <!-- Unread dot -->
          <div v-if="!n.read" class="w-2 h-2 rounded-full bg-primary flex-shrink-0 mt-1.5" />
        </div>

        <p v-if="notifications.length === 0" class="text-center text-muted-foreground text-sm font-body italic py-6">
          Sem notificações
        </p>
      </div>
    </div>
  </div>
</template>
