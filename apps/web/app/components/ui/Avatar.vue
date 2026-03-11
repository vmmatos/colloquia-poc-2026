<script setup lang="ts">
const props = withDefaults(defineProps<{
  name: string
  src?: string | null
  size?: 'sm' | 'md' | 'lg' | 'xl'
}>(), {
  src: null,
  size: 'md',
})

const sizeClasses: Record<string, string> = {
  sm: 'w-6 h-6 text-xs',
  md: 'w-8 h-8 text-sm',
  lg: 'w-10 h-10 text-base',
  xl: 'w-20 h-20 text-2xl',
}
</script>

<template>
  <img
    v-if="src"
    :src="src"
    :alt="name"
    :class="['rounded-full object-cover flex-shrink-0', sizeClasses[size]]"
  />
  <div
    v-else
    :class="['rounded-full bg-secondary border border-border text-foreground flex items-center justify-center font-heading font-medium flex-shrink-0', sizeClasses[size]]"
  >
    {{ name.charAt(0).toUpperCase() }}
  </div>
</template>
