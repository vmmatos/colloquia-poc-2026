<script setup lang="ts">
import type { UserProfile } from '../../shared/types/auth'

const props = defineProps<{
  open: boolean
  existingMemberIds: string[]
}>()
const emit = defineEmits<{ close: [] }>()

const { t } = useI18n()
const { createChannel } = useChannels()
const { auth } = useAuth()
const { resolveUser, prefetchUsers } = useUsersCache()
const config = useRuntimeConfig()

const groupName = ref('')
const memberQuery = ref('')
const memberResults = ref<UserProfile[]>([])
const extraMembers = ref<UserProfile[]>([])
const searchLoading = ref(false)
const loading = ref(false)
const error = ref<string | null>(null)
let searchTimer: ReturnType<typeof setTimeout> | null = null

// Prefetch display names for existing members
watch(() => props.open, async (open) => {
  if (open) {
    await prefetchUsers(props.existingMemberIds)
  }
})

watch(memberQuery, (q) => {
  if (searchTimer) clearTimeout(searchTimer)
  memberResults.value = []
  if (q.trim().length < 2) return
  searchTimer = setTimeout(async () => {
    searchLoading.value = true
    try {
      const res = await $fetch<UserProfile[]>(`${config.public.apiBase}/api/v1/users/search`, {
        query: { q: q.trim() },
        headers: { Authorization: `Bearer ${auth.value.access_token}` },
      })
      // Exclude already selected or existing members
      const allExcluded = new Set([
        ...props.existingMemberIds,
        ...extraMembers.value.map(m => m.user_id),
      ])
      memberResults.value = (res ?? []).filter(u => !allExcluded.has(u.user_id))
    } catch {
      memberResults.value = []
    } finally {
      searchLoading.value = false
    }
  }, 300)
})

function addMember(user: UserProfile) {
  if (!extraMembers.value.find(m => m.user_id === user.user_id)) {
    extraMembers.value = [...extraMembers.value, user]
  }
  memberQuery.value = ''
  memberResults.value = []
}

function removeMember(userId: string) {
  extraMembers.value = extraMembers.value.filter(m => m.user_id !== userId)
}

async function submit() {
  if (extraMembers.value.length === 0) {
    error.value = t('group.noMembersError')
    return
  }
  error.value = null
  loading.value = true
  try {
    const allMemberIds = [
      ...props.existingMemberIds,
      ...extraMembers.value.map(m => m.user_id),
    ]
    const ch = await createChannel({
      name: groupName.value.trim() || t('group.groupFallback'),
      is_private: true,
      type: 'group',
      initial_member_ids: allMemberIds,
    })
    close()
    await navigateTo(`/channels/${ch.id}`)
  } catch (err: unknown) {
    const msg = (err as { data?: { error?: string } })?.data?.error
    error.value = msg ?? t('group.error')
  } finally {
    loading.value = false
  }
}

function close() {
  groupName.value = ''
  memberQuery.value = ''
  memberResults.value = []
  extraMembers.value = []
  error.value = null
  emit('close')
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
            <h2 class="font-heading font-semibold text-foreground text-base">{{ $t('group.title') }}</h2>
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
            <!-- Existing members (read-only chips) -->
            <div>
              <label class="block text-xs font-heading font-medium text-muted-foreground uppercase tracking-wider mb-1.5">
                {{ $t('group.currentConversation') }}
              </label>
              <div class="flex flex-wrap gap-1.5">
                <span
                  v-for="id in existingMemberIds"
                  :key="id"
                  class="inline-flex items-center gap-1 px-2 py-0.5 bg-secondary/60 rounded-full text-xs font-heading text-muted-foreground"
                >
                  {{ resolveUser(id) }}
                </span>
              </div>
            </div>

            <!-- Group name (optional) -->
            <div>
              <label class="block text-xs font-heading font-medium text-muted-foreground uppercase tracking-wider mb-1.5">
                {{ $t('group.groupNameLabel') }} <span class="text-muted-foreground font-normal">{{ $t('group.groupNameOptional') }}</span>
              </label>
              <input
                v-model="groupName"
                type="text"
                :placeholder="$t('group.groupNamePlaceholder')"
                maxlength="80"
                class="w-full bg-background border border-border rounded-md px-3 py-2 text-sm font-heading text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary/50 transition-colors"
              />
            </div>

            <!-- Add new members -->
            <div>
              <label class="block text-xs font-heading font-medium text-muted-foreground uppercase tracking-wider mb-1.5">
                {{ $t('group.addPeopleLabel') }} <span class="text-destructive">*</span>
              </label>

              <!-- Selected extra members -->
              <div v-if="extraMembers.length > 0" class="flex flex-wrap gap-1.5 mb-2">
                <span
                  v-for="m in extraMembers"
                  :key="m.user_id"
                  class="inline-flex items-center gap-1 px-2 py-0.5 bg-secondary rounded-full text-xs font-heading text-foreground"
                >
                  {{ m.name || m.user_id }}
                  <button
                    type="button"
                    class="text-muted-foreground hover:text-foreground ml-0.5"
                    @click="removeMember(m.user_id)"
                  >
                    <svg class="h-3 w-3" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </span>
              </div>

              <div class="relative">
                <input
                  v-model="memberQuery"
                  type="text"
                  :placeholder="$t('group.searchPlaceholder')"
                  class="w-full bg-background border border-border rounded-md px-3 py-2 text-sm font-heading text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary/50 transition-colors"
                />
                <ul
                  v-if="memberResults.length > 0"
                  class="absolute z-10 mt-1 w-full bg-card border border-border rounded-md shadow-lg max-h-40 overflow-y-auto"
                >
                  <li
                    v-for="u in memberResults"
                    :key="u.user_id"
                    class="flex items-center gap-2 px-3 py-2 cursor-pointer hover:bg-secondary/50 text-sm font-heading text-foreground"
                    @mousedown.prevent="addMember(u)"
                  >
                    <UiAvatar :name="u.name || u.user_id" size="sm" />
                    <span>{{ u.name || u.user_id }}</span>
                  </li>
                </ul>
              </div>
            </div>

            <p v-if="error" class="text-xs text-destructive font-body">{{ error }}</p>
          </div>

          <!-- Footer -->
          <div class="flex justify-end gap-3 px-6 py-4 border-t border-border">
            <button
              class="px-4 py-2 text-sm font-heading text-muted-foreground hover:text-foreground transition-colors"
              @click="close"
            >
              {{ $t('common.cancel') }}
            </button>
            <UiButton :loading="loading" @click="submit">
              {{ $t('group.submit') }}
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
