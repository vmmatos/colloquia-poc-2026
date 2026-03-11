<script setup lang="ts">
definePageMeta({ middleware: 'guest', layout: 'auth' })

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
    error.value = e?.data?.message || e?.message || 'Falha no login. Tenta novamente.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <UiCard>
    <form class="space-y-4" @submit.prevent="handleSubmit">
      <div class="mb-2">
        <h2 class="text-lg font-heading font-semibold text-foreground">Entrar</h2>
        <p class="text-xs font-body text-muted-foreground mt-0.5">Bem-vindo de volta</p>
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

      <div class="h-4">
        <p v-if="error" class="text-destructive text-xs font-heading leading-4">{{ error }}</p>
      </div>

      <UiButton type="submit" :loading="loading" class="w-full">
        Entrar
      </UiButton>
    </form>

    <p class="mt-4 text-center text-xs font-heading text-muted-foreground">
      Ainda não tens conta?
      <NuxtLink to="/register" class="text-primary hover:text-primary/80 transition-colors">
        Registar
      </NuxtLink>
    </p>
  </UiCard>
</template>
