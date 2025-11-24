import { reactive, computed } from 'vue'

// Decode JWT token helper
function decodeToken(token) {
  try {
    const base64Url = token.split('.')[1]
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    )
    return JSON.parse(jsonPayload)
  } catch (error) {
    return null
  }
}

const state = reactive({
  isAuthenticated: !!localStorage.getItem('jwt_token'),
  username: localStorage.getItem('username') || ''
})

export const authStore = {
  state,
  
  isAuthenticated: computed(() => state.isAuthenticated),
  username: computed(() => state.username),
  
  login(token, username) {
    localStorage.setItem('jwt_token', token)
    localStorage.setItem('username', username)
    state.isAuthenticated = true
    state.username = username
  },
  
  logout() {
    localStorage.removeItem('jwt_token')
    localStorage.removeItem('username')
    state.isAuthenticated = false
    state.username = ''
  },
  
  checkAuth() {
    const token = localStorage.getItem('jwt_token')
    const username = localStorage.getItem('username')
    state.isAuthenticated = !!token
    state.username = username || ''
  },

  getUserId() {
    const token = localStorage.getItem('jwt_token')
    if (!token) return null
    const decoded = decodeToken(token)
    return decoded?.uid || null
  },

  getUserRoles() {
    const token = localStorage.getItem('jwt_token')
    if (!token) return []
    const decoded = decodeToken(token)
    return decoded?.roles || []
  },

  isGuide() {
    const roles = this.getUserRoles()
    return roles.includes('guide') || roles.includes('Guide')
  }
}
