<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '~/composables/useAuth'

const router = useRouter()
const auth = useAuth()

onMounted(() => {
  auth.loadUserFromStorage()
})

function handleLogout() {
  auth.clearUser()
  router.push('/auth/login')
}

function handleLogin() {
  router.push('/auth/login')
}
</script>

<template>
  <div>

    <nav class="flex items-center justify-between gap-6 p-4 border-b bg-white shadow-sm">

      <!-- Left side navigation -->
      <div class="flex gap-6">
        <NuxtLink to="/" class="font-medium hover:text-blue-600 transition-colors">Home</NuxtLink>
        <NuxtLink to="/teams" class="font-medium hover:text-blue-600 transition-colors">Teams</NuxtLink>
        <NuxtLink to="/leagues" class="font-medium hover:text-blue-600 transition-colors">Leagues</NuxtLink>
        <NuxtLink v-if="auth.isAuthenticated.value" to="/stats" class="font-medium hover:text-blue-600 transition-colors">Stats</NuxtLink>
        <NuxtLink v-if="auth.isAuthenticated.value" to="/history" class="font-medium hover:text-blue-600 transition-colors">History</NuxtLink>
      </div>

      <!-- Right side - Login/User menu -->
      <div class="flex items-center gap-4">

        <!-- Logged Out State -->
        <button v-if="!auth.isAuthenticated.value" @click="handleLogin"
          class="px-4 py-2 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700 transition-colors">
          Login
        </button>

        <!-- Logged In State -->
        <div v-else class="flex items-center gap-3">
          <div class="text-right">
            <p class="text-sm font-semibold text-gray-900">
              {{ auth.currentUser.value?.first_name }} {{ auth.currentUser.value?.last_name }}
            </p>
            <p class="text-xs text-gray-500">
              {{ auth.currentUser.value?.roles.join(', ') }}
            </p>
          </div>
          <button @click="handleLogout"
            class="px-4 py-2 text-gray-700 border border-gray-300 rounded-lg font-medium hover:bg-gray-50 transition-colors">
            Logout
          </button>
        </div>

      </div>

    </nav>

    <main class="p-6">
      <slot />
    </main>

  </div>
</template>