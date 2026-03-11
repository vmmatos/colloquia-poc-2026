<script setup lang="ts">
definePageMeta({ middleware: 'guest', layout: 'auth' })

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
    confirmError.value = 'As palavras-passe não coincidem.'
    return
  }

  loading.value = true
  try {
    await register(email.value, password.value)
    await navigateTo('/')
  } catch (e: any) {
    error.value = e?.data?.message || e?.message || 'Falha no registo. Tenta novamente.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <UiCard>
    <form class="space-y-4" @submit.prevent="handleSubmit">
      <div class="mb-2">
        <h2 class="text-lg font-heading font-semibold text-foreground">Criar conta</h2>
        <p class="text-xs font-body text-muted-foreground mt-0.5">Junta-te ao Colloquia</p>
      </div>

      <UiInput
        v-model="email"
        type="email"
        placeholder="Email"
        :error="null"
      />
      <UiInput
        v-model="password"
        type="password"
        placeholder="Palavra-passe"
        :error="null"
      />
      <UiInput
        v-model="confirmPassword"
        type="password"
        placeholder="Confirmar palavra-passe"
        :error="confirmError || null"
      />

      <div class="h-4">
        <p v-if="error" class="text-destructive text-xs font-heading leading-4">{{ error }}</p>
      </div>

      <UiButton type="submit" :loading="loading" class="w-full">
        Criar conta
      </UiButton>
    </form>

    <p class="mt-4 text-center text-xs font-heading text-muted-foreground">
      Já tens conta?
      <NuxtLink to="/login" class="text-primary hover:text-primary/80 transition-colors">
        Entrar
      </NuxtLink>
    </p>
  </UiCard>
</template>
