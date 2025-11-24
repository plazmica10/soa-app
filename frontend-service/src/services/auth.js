import { api } from './api'
import { authStore } from '../stores/authStore'

export const authService = {
  async login(credentials) {
    try {
      const response = await api.login(credentials)
      
      // Store JWT token - handle both token and access_token
      const token = response.token || response.access_token
      if (token) {
        authStore.login(token, credentials.username)
      }
      
      return response
    } catch (error) {
      const message = error.response?.data?.message || 
                     error.response?.data || 
                     error.message ||
                     'Login failed'
      throw new Error(typeof message === 'string' ? message : 'Login failed')
    }
  },

  async register(userData) {
    try {
      const response = await api.register(userData)
      return response
    } catch (error) {
      // Parse detailed error messages
      let message = 'Registration failed'
      
      if (error.response?.data) {
        const errorData = error.response.data
        
        // Handle string error messages
        if (typeof errorData === 'string') {
          message = errorData
          
          // Make error messages more user-friendly
          if (errorData.includes('user exists') || errorData.includes('duplicate')) {
            if (errorData.includes('username')) {
              message = 'Username already exists. Please choose a different username.'
            } else if (errorData.includes('email')) {
              message = 'Email already registered. Please use a different email.'
            } else {
              message = 'User already exists. Username or email may already be registered.'
            }
          } else if (errorData.includes('invalid')) {
            message = 'Invalid input. Please check your information.'
          }
        } 
        // Handle object error messages
        else if (errorData.message) {
          message = errorData.message
        }
      } else if (error.message) {
        message = error.message
      }
      
      throw new Error(message)
    }
  },

  logout() {
    authStore.logout()
  },

  isAuthenticated() {
    return authStore.isAuthenticated.value
  },

  getToken() {
    return localStorage.getItem('jwt_token')
  },

  getUsername() {
    return authStore.username.value
  },

  // Decode JWT token (simple implementation)
  decodeToken(token) {
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
  },

  getUserFromToken() {
    const token = this.getToken()
    if (!token) return null
    return this.decodeToken(token)
  }
}
