<template>
  <div class="min-h-screen bg-gray-950 text-gray-100 p-6">
    <h1 class="text-2xl font-bold mb-6 text-white">Auth Debug Page</h1>

    <!-- Auth State Panel -->
    <div class="bg-gray-900 rounded-lg p-4 mb-6 border border-gray-800">
      <h2 class="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-3">Auth State</h2>
      <div class="flex items-center gap-3 mb-2">
        <span
          class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
          :class="isAuthenticated ? 'bg-green-900 text-green-300' : 'bg-red-900 text-red-300'"
        >
          {{ isAuthenticated ? 'Authenticated' : 'Unauthenticated' }}
        </span>
        <span v-if="tokenExpiresIn" class="text-xs text-gray-500">
          expires in: <span class="text-yellow-400">{{ tokenExpiresIn }}</span>
        </span>
      </div>
      <div v-if="auth.user_id" class="text-xs text-gray-400 space-y-1 font-mono">
        <div>user_id: <span class="text-blue-400">{{ auth.user_id }}</span></div>
        <div v-if="auth.access_token">
          token:
          <span class="text-purple-400">
            {{ auth.access_token.slice(0, 20) }}...{{ auth.access_token.slice(-8) }}
          </span>
        </div>
      </div>
    </div>

    <!-- Register + Login Forms -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
      <!-- Register -->
      <div class="bg-gray-900 rounded-lg p-4 border border-gray-800">
        <h2 class="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-3">Register</h2>
        <div class="space-y-2">
          <input
            v-model="registerEmail"
            type="email"
            placeholder="email"
            class="w-full bg-gray-800 border border-gray-700 rounded px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-blue-500"
          />
          <input
            v-model="registerPassword"
            type="password"
            placeholder="password"
            class="w-full bg-gray-800 border border-gray-700 rounded px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-blue-500"
          />
          <button
            :disabled="loadingRegister"
            class="w-full bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium py-2 px-4 rounded transition-colors"
            :class="{ 'opacity-50 cursor-not-allowed': loadingRegister }"
            @click="handleRegister"
          >
            {{ loadingRegister ? 'Registering...' : 'Register' }}
          </button>
        </div>
      </div>

      <!-- Login -->
      <div class="bg-gray-900 rounded-lg p-4 border border-gray-800">
        <h2 class="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-3">Login</h2>
        <div class="space-y-2">
          <input
            v-model="loginEmail"
            type="email"
            placeholder="email"
            class="w-full bg-gray-800 border border-gray-700 rounded px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-blue-500"
          />
          <input
            v-model="loginPassword"
            type="password"
            placeholder="password"
            class="w-full bg-gray-800 border border-gray-700 rounded px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-blue-500"
          />
          <button
            :disabled="loadingLogin"
            class="w-full bg-green-600 hover:bg-green-700 text-white text-sm font-medium py-2 px-4 rounded transition-colors"
            :class="{ 'opacity-50 cursor-not-allowed': loadingLogin }"
            @click="handleLogin"
          >
            {{ loadingLogin ? 'Logging in...' : 'Login' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Action Buttons -->
    <div class="flex flex-wrap gap-3 mb-6">
      <button
        :disabled="loadingValidate"
        class="bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium py-2 px-4 rounded transition-colors"
        :class="{ 'opacity-50 cursor-not-allowed': loadingValidate }"
        @click="handleValidate"
      >
        {{ loadingValidate ? 'Validating...' : 'Validate Token' }}
      </button>

      <button
        :disabled="loadingProfile"
        class="bg-teal-600 hover:bg-teal-700 text-white text-sm font-medium py-2 px-4 rounded transition-colors"
        :class="{ 'opacity-50 cursor-not-allowed': loadingProfile }"
        @click="handleGetProfile"
      >
        {{ loadingProfile ? 'Loading...' : 'Get Profile' }}
      </button>

      <button
        :disabled="loadingRefresh"
        class="bg-yellow-600 hover:bg-yellow-700 text-white text-sm font-medium py-2 px-4 rounded transition-colors"
        :class="{ 'opacity-50 cursor-not-allowed': loadingRefresh }"
        @click="handleRefresh"
      >
        {{ loadingRefresh ? 'Refreshing...' : 'Refresh Token' }}
      </button>

      <button
        :disabled="loadingLogout"
        class="bg-red-600 hover:bg-red-700 text-white text-sm font-medium py-2 px-4 rounded transition-colors"
        :class="{ 'opacity-50 cursor-not-allowed': loadingLogout }"
        @click="handleLogout"
      >
        {{ loadingLogout ? 'Logging out...' : 'Logout' }}
      </button>
    </div>

    <!-- Response Display -->
    <div class="bg-gray-900 rounded-lg p-4 border border-gray-800">
      <h2 class="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-3">Last Response</h2>
      <pre
        class="text-xs font-mono rounded p-3 overflow-auto max-h-96"
        :class="hasError ? 'bg-red-950 text-red-300' : 'bg-gray-950 text-green-400'"
      >
        {{ lastResponseText }}
      </pre>
    </div>
  </div>
</template>

<script setup lang="ts">
const {
  auth,
  isAuthenticated,
  tokenExpiresIn,
  register,
  login,
  logout,
  refreshToken,
  getProfile,
  validateToken
} = useAuth()

const registerEmail = ref('')
const registerPassword = ref('')
const loginEmail = ref('')
const loginPassword = ref('')

const loadingRegister = ref(false)
const loadingLogin = ref(false)
const loadingValidate = ref(false)
const loadingProfile = ref(false)
const loadingRefresh = ref(false)
const loadingLogout = ref(false)

const lastResponse = ref<unknown>(null)
const hasError = ref(false)

const lastResponseText = computed(() =>
  lastResponse.value === null
    ? '// No response yet'
    : JSON.stringify(lastResponse.value, null, 2)
)

async function withLoading(flag: Ref<boolean>, fn: () => Promise<unknown>) {
  flag.value = true
  hasError.value = false
  try {
    const result = await fn()
    lastResponse.value = result
  } catch (err: unknown) {
    hasError.value = true
    lastResponse.value = {
      error: (err as Error).message,
      data: (err as { data?: unknown }).data,
    }
  } finally {
    flag.value = false
  }
}

const handleRegister = () =>
  withLoading(loadingRegister, () => register(registerEmail.value, registerPassword.value))

const handleLogin = () =>
  withLoading(loadingLogin, () => login(loginEmail.value, loginPassword.value))

const handleValidate = () => withLoading(loadingValidate, validateToken)
const handleGetProfile = () => withLoading(loadingProfile, getProfile)
const handleRefresh = () => withLoading(loadingRefresh, refreshToken)
const handleLogout = () => withLoading(loadingLogout, logout)
</script>
