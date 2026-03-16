<script setup lang="ts">
definePageMeta({ middleware: 'auth' })

const { channels, fetchMyChannels } = useChannels()
const openSidebar = inject<() => void>('openSidebar')

onMounted(async () => {
  if (channels.value.length === 0) {
    try { await fetchMyChannels() } catch { /* ignore */ }
  }
  if (channels.value.length > 0) {
    await navigateTo(`/channels/${channels?.value[0]?.id}`, { replace: true })
  }
})
</script>

<template>
  <div class="flex flex-col h-full bg-background">
    <!-- Header -->
    <header class="h-14 flex items-center justify-between px-4 md:px-6 border-b border-border flex-shrink-0">
      <div class="flex items-center gap-2">
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
        <span class="font-heading font-semibold text-foreground text-sm">Colloquia</span>
      </div>
      <NotificationCenter />
    </header>

    <!-- Empty state -->
    <div class="flex-1 flex flex-col items-center justify-center gap-4 px-4 text-center">
      <div class="w-12 h-12 rounded-xl bg-secondary flex items-center justify-center">
        <svg class="h-6 w-6 text-muted-foreground" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
        </svg>
      </div>
      <div>
        <h2 class="font-heading font-semibold text-foreground text-base">Sem canais ainda</h2>
        <p class="text-sm font-body text-muted-foreground mt-1">
          Cria o teu primeiro canal para começar a colaborar.
        </p>
      </div>
    </div>
  </div>
</template>
