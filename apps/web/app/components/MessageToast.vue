<script setup lang="ts">
export interface Toast {
  id: number
  userId: string
  author: string
  preview: string
  channel: string
}

defineProps<{ toasts: Toast[] }>()
const emit = defineEmits<{ dismiss: [id: number] }>()

const { isOnline } = usePresence()
</script>

<template>
  <div class="fixed bottom-4 right-4 z-50 flex flex-col gap-2 pointer-events-none">
    <TransitionGroup name="toast">
      <div
        v-for="t in toasts"
        :key="t.id"
        class="bg-card border border-border rounded-lg shadow-xl p-3 w-72 pointer-events-auto animate-fade-in"
      >
        <div class="flex items-start gap-2">
          <UiAvatar :name="t.author" size="sm" :online="isOnline(t.userId)" />
          <div class="flex-1 min-w-0">
            <p class="text-xs font-heading font-semibold text-foreground">{{ t.author }}</p>
            <p class="text-xs font-body text-muted-foreground truncate">{{ t.preview }}</p>
            <p class="text-xs text-muted-foreground/60 mt-0.5">{{ t.channel }}</p>
          </div>
          <button
            class="text-muted-foreground hover:text-foreground flex-shrink-0 text-base leading-none"
            @click="emit('dismiss', t.id)"
          >
            ×
          </button>
        </div>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.2s ease;
}
.toast-enter-from {
  opacity: 0;
  transform: translateY(8px);
}
.toast-leave-to {
  opacity: 0;
  transform: translateX(100%);
}
</style>
