<script setup lang="ts">
import type { CreateChannelInput } from '../../shared/types/channels'

defineProps<{ open: boolean }>()
const emit = defineEmits<{ close: [] }>()

const { createChannel } = useChannels()

const name = ref('')
const description = ref('')
const isPrivate = ref(false)
const loading = ref(false)
const error = ref<string | null>(null)

function reset() {
  name.value = ''
  description.value = ''
  isPrivate.value = false
  error.value = null
}

function close() {
  reset()
  emit('close')
}

async function submit() {
  error.value = null
  const trimmedName = name.value.trim()
  if (!trimmedName) {
    error.value = 'O nome do canal é obrigatório.'
    return
  }

  loading.value = true
  try {
    const input: CreateChannelInput = {
      name: trimmedName,
      description: description.value.trim() || undefined,
      is_private: isPrivate.value,
    }
    const ch = await createChannel(input)
    close()
    await navigateTo(`/channels/${ch.id}`)
  } catch (err: unknown) {
    const msg = (err as { data?: { error?: string } })?.data?.error
    error.value = msg ?? 'Erro ao criar o canal. Tenta novamente.'
  } finally {
    loading.value = false
  }
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
        <div class="relative z-10 w-full max-w-md bg-card border border-border rounded-lg shadow-xl">
          <!-- Header -->
          <div class="flex items-center justify-between px-6 py-4 border-b border-border">
            <h2 class="font-heading font-semibold text-foreground text-base">Criar canal</h2>
            <button
              class="text-muted-foreground hover:text-foreground transition-colors"
              @click="close"
            >
              <svg class="h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- Body -->
          <div class="px-6 py-5 space-y-4">
            <!-- Name -->
            <div>
              <label class="block text-xs font-heading font-medium text-muted-foreground uppercase tracking-wider mb-1.5">
                Nome <span class="text-destructive">*</span>
              </label>
              <input
                v-model="name"
                type="text"
                placeholder="ex: marketing, anúncios"
                maxlength="80"
                class="w-full bg-background border border-border rounded-md px-3 py-2 text-sm font-heading text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary/50 transition-colors"
                @keydown.enter="submit"
              />
            </div>

            <!-- Description -->
            <div>
              <label class="block text-xs font-heading font-medium text-muted-foreground uppercase tracking-wider mb-1.5">
                Descrição <span class="text-muted-foreground font-normal">(opcional)</span>
              </label>
              <input
                v-model="description"
                type="text"
                placeholder="Para que serve este canal?"
                maxlength="255"
                class="w-full bg-background border border-border rounded-md px-3 py-2 text-sm font-heading text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary/50 transition-colors"
              />
            </div>

            <!-- Private toggle -->
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm font-heading text-foreground">Canal privado</p>
                <p class="text-xs font-body text-muted-foreground mt-0.5">Só quem for adicionado pode ver e entrar</p>
              </div>
              <button
                :class="[
                  'relative inline-flex h-6 w-11 flex-shrink-0 items-center rounded-full transition-colors',
                  isPrivate ? 'bg-primary' : 'bg-secondary',
                ]"
                @click="isPrivate = !isPrivate"
              >
                <span
                  :class="[
                    'inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform',
                    isPrivate ? 'translate-x-6' : 'translate-x-1',
                  ]"
                />
              </button>
            </div>

            <!-- Error -->
            <p v-if="error" class="text-xs text-destructive font-body">{{ error }}</p>
          </div>

          <!-- Footer -->
          <div class="flex justify-end gap-3 px-6 py-4 border-t border-border">
            <button
              class="px-4 py-2 text-sm font-heading text-muted-foreground hover:text-foreground transition-colors"
              @click="close"
            >
              Cancelar
            </button>
            <UiButton :loading="loading" @click="submit">
              Criar canal
            </UiButton>
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
