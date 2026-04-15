<script setup lang="ts">
definePageMeta({ middleware: 'guest', layout: 'auth' })

const { t } = useI18n()
const { register } = useAuth()

const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const error = ref('')
const confirmError = ref('')

async function handleSubmit() {
  error.value = ''
  confirmError.value = ''

  if (password.value !== confirmPassword.value) {
    confirmError.value = t('auth.register.passwordMismatch')
    return
  }

  loading.value = true
  try {
    await register(email.value, password.value)
    await navigateTo('/')
  } catch (e: any) {
    error.value = e?.data?.message || e?.message || t('auth.register.error')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <UiCard>
    <form class="space-y-4" @submit.prevent="handleSubmit">
      <div class="mb-2">
        <h2 class="text-lg font-heading font-semibold text-foreground">{{ $t('auth.register.title') }}</h2>
        <p class="text-xs font-body text-muted-foreground mt-0.5">{{ $t('auth.register.subtitle') }}</p>
      </div>

      <UiInput
        v-model="email"
        type="email"
        :placeholder="$t('auth.register.emailPlaceholder')"
        :error="null"
      />
      <UiInput
        v-model="password"
        type="password"
        :placeholder="$t('auth.register.passwordPlaceholder')"
        :error="null"
      />
      <UiInput
        v-model="confirmPassword"
        type="password"
        :placeholder="$t('auth.register.confirmPlaceholder')"
        :error="confirmError || null"
      />

      <div class="h-4">
        <p v-if="error" class="text-destructive text-xs font-heading leading-4">{{ error }}</p>
      </div>

      <UiButton type="submit" :loading="loading" class="w-full">
        {{ $t('auth.register.submit') }}
      </UiButton>
    </form>

    <p class="mt-4 text-center text-xs font-heading text-muted-foreground">
      {{ $t('auth.register.hasAccount') }}
      <NuxtLink to="/login" class="text-primary hover:text-primary/80 transition-colors">
        {{ $t('auth.register.login') }}
      </NuxtLink>
    </p>
  </UiCard>
</template>
