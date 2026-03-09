<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const isLoggedIn = ref(false)
const username = ref('')

onMounted(() => {
  // Check if user is logged in
  const authToken = localStorage.getItem('authToken')
  const storedUsername = localStorage.getItem('username')

  if (authToken) {
    isLoggedIn.value = true
    username.value = storedUsername || 'User'
  }
})

function handleLogout() {
  localStorage.removeItem('authToken')
  localStorage.removeItem('username')
  isLoggedIn.value = false
  username.value = ''
  router.push('/auth/login')
}

function handleLogin() {
  router.push('/auth/login')
}
</script>

<template>
  <div>

    <nav class="flex items-center justify-between gap-6 p-4 border-b">

      <!-- Left side navigation -->
      <div class="flex gap-6">
        <NuxtLink to="/" class="hover:text-blue-600 transition-colors">Home</NuxtLink>
        <NuxtLink to="/leagues" class="hover:text-blue-600 transition-colors">Leagues</NuxtLink>
        <NuxtLink to="/stats" class="hover:text-blue-600 transition-colors">Stats</NuxtLink>
      </div>

      <!-- Right side - Login/User menu -->
      <div class="flex items-center gap-4">

        <!-- Logged Out State -->
        <button v-if="!isLoggedIn" @click="handleLogin"
          class="px-4 py-2 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700 transition-colors">
          Login
        </button>

        <!-- Logged In State -->
        <div v-else class="flex items-center gap-3">
          <span class="text-sm text-gray-700">
            Welcome, <span class="font-semibold">{{ username }}</span>
          </span>
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