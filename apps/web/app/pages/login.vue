<script setup lang="ts">
definePageMeta({ middleware: 'guest', layout: 'auth' })

const { t } = useI18n()
const { login } = useAuth()

const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

async function handleSubmit() {
  error.value = ''
  loading.value = true
  try {
    await login(email.value, password.value)
    await navigateTo('/')
  } catch (e: any) {
    const status = e?.status ?? e?.response?.status
    if (status === 401) {
      error.value = t('auth.login.invalidCredentials')
    } else if (status === 423) {
      error.value = t('auth.login.accountLocked')
    } else {
      error.value = t('auth.login.error')
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <UiCard>
    <form class="space-y-4" @submit.prevent="handleSubmit">
      <div class="mb-2">
        <h2 class="text-lg font-heading font-semibold text-foreground">{{ $t('auth.login.title') }}</h2>
        <p class="text-xs font-body text-muted-foreground mt-0.5">{{ $t('auth.login.subtitle') }}</p>
      </div>

      <UiInput
        v-model="email"
        type="email"
        :placeholder="$t('auth.login.emailPlaceholder')"
        :error="null"
      />
      <UiInput
        v-model="password"
        type="password"
        :placeholder="$t('auth.login.passwordPlaceholder')"
        :error="null"
      />

      <div class="h-4">
        <p v-if="error" class="text-destructive text-xs font-heading leading-4">{{ error }}</p>
      </div>

      <UiButton type="submit" :loading="loading" class="w-full">
        {{ $t('auth.login.submit') }}
      </UiButton>
    </form>

    <p class="mt-4 text-center text-xs font-heading text-muted-foreground">
      {{ $t('auth.login.noAccount') }}
      <NuxtLink to="/register" class="text-primary hover:text-primary/80 transition-colors">
        {{ $t('auth.login.register') }}
      </NuxtLink>
    </p>
  </UiCard>
</template>
