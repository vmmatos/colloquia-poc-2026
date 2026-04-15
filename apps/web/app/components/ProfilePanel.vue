<script setup lang="ts">
const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const { t } = useI18n()
const { auth, getProfile, refreshToken } = useAuth()
const { currentLocale, availableLocales, setLocale } = useLocale()

const displayName = ref('')
const bio = ref('')
const loading = ref(false)
const saving = ref(false)
const error = ref('')

// Load profile when panel opens
watch(() => props.open, async (val) => {
  if (!val) return
  error.value = ''
  loading.value = true
  try {
    const profile = await getProfile()
    displayName.value = profile.name || ''
    bio.value = profile.bio || ''
  } catch {
    error.value = t('profile.loadError')
  } finally {
    loading.value = false
  }
})

async function patchProfile() {
  await $fetch('/api/users/me', {
    method: 'PATCH',
    headers: { Authorization: `Bearer ${auth.value.access_token}` },
    body: { name: displayName.value, bio: bio.value },
  })
}

async function saveProfile() {
  error.value = ''
  saving.value = true
  try {
    await patchProfile()
    emit('close')
  } catch (e: any) {
    if (e?.status === 401 || e?.response?.status === 401) {
      try {
        await refreshToken()
        await patchProfile()
        emit('close')
      } catch {
        error.value = t('profile.sessionExpired')
      }
    } else {
      error.value = e?.data?.message || e?.message || t('profile.saveError')
    }
  } finally {
    saving.value = false
  }
}

const initials = computed(() => {
  const n = displayName.value || auth.value.user_id || 'U'
  return n.charAt(0).toUpperCase()
})
</script>

<template>
  <Teleport to="body">
    <Transition name="profile-panel">
      <div v-if="open" class="fixed inset-0 z-40">
        <!-- Backdrop -->
        <div
          class="absolute inset-0 bg-background/60 backdrop-blur-sm"
          @click="emit('close')"
        />

        <!-- Panel -->
        <div class="absolute right-0 top-0 h-full w-full sm:w-80 bg-surface-overlay/95 backdrop-blur-md border-l border-border z-50 flex flex-col animate-slide-in">
          <!-- Header -->
          <div class="h-14 flex items-center justify-between px-5 border-b border-border flex-shrink-0">
            <span class="text-sm font-heading font-semibold text-foreground">{{ $t('profile.title') }}</span>
            <button
              class="p-1.5 rounded-md text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-colors"
              @click="emit('close')"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- Content -->
          <div class="flex-1 overflow-y-auto py-6 px-5 space-y-6">
            <div v-if="loading" class="flex items-center justify-center py-12">
              <span class="text-muted-foreground text-sm font-body italic">{{ $t('common.loading') }}</span>
            </div>

            <template v-else>
              <!-- Avatar -->
              <div class="flex justify-center">
                <div class="relative group">
                  <div class="w-20 h-20 rounded-full bg-secondary border border-border flex items-center justify-center text-2xl font-heading font-medium text-foreground select-none">
                    {{ initials }}
                  </div>
                </div>
              </div>

              <!-- Name -->
              <div>
                <label class="block text-xs font-heading text-muted-foreground mb-1.5 uppercase tracking-wider">
                  {{ $t('profile.displayName') }}
                </label>
                <input
                  v-model="displayName"
                  type="text"
                  :placeholder="$t('profile.displayNamePlaceholder')"
                  class="profile-input"
                />
              </div>

              <!-- Bio -->
              <div>
                <label class="block text-xs font-heading text-muted-foreground mb-1.5 uppercase tracking-wider">
                  {{ $t('profile.bio') }}
                </label>
                <textarea
                  v-model="bio"
                  rows="3"
                  :placeholder="$t('profile.bioPlaceholder')"
                  class="profile-input resize-none font-body italic"
                />
              </div>

              <!-- Language selector -->
              <div>
                <label class="block text-xs font-heading text-muted-foreground mb-1.5 uppercase tracking-wider">
                  {{ $t('profile.language') }}
                </label>
                <select
                  :value="currentLocale"
                  class="profile-input"
                  @change="setLocale(($event.target as HTMLSelectElement).value)"
                >
                  <option
                    v-for="loc in availableLocales"
                    :key="loc.code"
                    :value="loc.code"
                  >
                    {{ loc.name }}
                  </option>
                </select>
              </div>

              <!-- User ID (read-only) -->
              <div class="pt-2 border-t border-border">
                <label class="block text-xs font-heading text-muted-foreground mb-1.5 uppercase tracking-wider">
                  {{ $t('profile.userId') }}
                </label>
                <p class="text-sm font-heading text-muted-foreground truncate">{{ auth.user_id || '—' }}</p>
              </div>

              <p v-if="error" class="text-destructive text-xs font-heading">{{ error }}</p>
            </template>
          </div>

          <!-- Footer -->
          <div class="flex-shrink-0 px-5 py-4 border-t border-border">
            <button
              :disabled="saving || loading"
              class="w-full bg-primary text-primary-foreground rounded-md py-2 text-sm font-heading font-medium hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              @click="saveProfile"
            >
              {{ saving ? $t('common.saving') : $t('common.save') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.profile-panel-enter-active,
.profile-panel-leave-active {
  transition: opacity 0.2s ease;
}
.profile-panel-enter-from,
.profile-panel-leave-to {
  opacity: 0;
}
</style>
