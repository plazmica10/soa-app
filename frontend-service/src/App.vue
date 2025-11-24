<template>
  <div id="app">
    <Navbar :key="navbarKey" />
    <router-view />
  </div>
</template>

<script>
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import Navbar from './components/Navbar.vue'
import { authStore } from './stores/authStore'

export default {
  name: 'App',
  components: {
    Navbar
  },
  setup() {
    const router = useRouter()
    const navbarKey = ref(0)

    // Watch for auth state changes and force navbar re-render
    watch(() => authStore.state.isAuthenticated, () => {
      navbarKey.value++
    })

    // Check auth on mount
    authStore.checkAuth()

    return {
      navbarKey
    }
  }
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
  background: #f5f5f5;
}

#app {
  min-height: 100vh;
}
</style>