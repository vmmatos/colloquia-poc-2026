<script setup lang="ts">
defineOptions({ inheritAttrs: false })

const props = withDefaults(defineProps<{
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger'
  loading?: boolean
  disabled?: boolean
}>(), {
  variant: 'primary',
  loading: false,
  disabled: false,
})

const variantClasses: Record<string, string> = {
  primary: 'bg-primary text-primary-foreground hover:bg-primary/90',
  danger: 'bg-destructive text-foreground hover:bg-destructive/90',
  secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary/80',
  ghost: 'bg-transparent text-muted-foreground hover:text-foreground hover:bg-secondary/50',
}
</script>

<template>
  <button
    v-bind="$attrs"
    :disabled="loading || disabled"
    :class="[
      'px-4 py-2 rounded-md font-heading font-medium text-sm transition-colors',
      variantClasses[variant],
      (loading || disabled) ? 'opacity-50 cursor-not-allowed' : '',
    ]"
  >
    <span v-if="loading">A carregar...</span>
    <slot v-else />
  </button>
</template>
