<script setup lang="ts">
const props = withDefaults(defineProps<{
  name: string
  src?: string | null
  size?: 'sm' | 'md' | 'lg' | 'xl'
  online?: boolean
}>(), {
  src: null,
  size: 'md',
  online: undefined,
})

const sizeClasses: Record<string, string> = {
  sm: 'w-6 h-6 text-xs',
  md: 'w-8 h-8 text-sm',
  lg: 'w-10 h-10 text-base',
  xl: 'w-20 h-20 text-2xl',
}

const dotClasses: Record<string, string> = {
  sm: 'w-2 h-2 bottom-0 right-0',
  md: 'w-2.5 h-2.5 bottom-0 right-0',
  lg: 'w-3 h-3 bottom-0.5 right-0.5',
  xl: 'w-4 h-4 bottom-1 right-1',
}
</script>

<template>
  <div class="relative inline-flex flex-shrink-0">
    <img
      v-if="src"
      :src="src"
      :alt="name"
      :class="['rounded-full object-cover', sizeClasses[size]]"
    />
    <div
      v-else
      :class="['rounded-full bg-secondary border border-border text-foreground flex items-center justify-center font-heading font-medium', sizeClasses[size]]"
    >
      {{ name.charAt(0).toUpperCase() }}
    </div>

    <!-- Presence indicator: only rendered when online prop is explicitly provided -->
    <span
      v-if="online !== undefined"
      :class="['absolute rounded-full border-2 border-background', dotClasses[size], online ? 'bg-green-500' : 'bg-zinc-500']"
    />
  </div>
</template>
