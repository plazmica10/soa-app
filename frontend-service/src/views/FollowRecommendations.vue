<template>
  <div class="recommendations">
    <h1>Discover People</h1>
    <p class="subtitle">Find interesting people to follow</p>

    <!-- Search Bar -->
    <div class="search-bar">
      <input 
        v-model="searchQuery" 
        type="text" 
        placeholder="Search by name or username..."
        @input="filterUsers"
        class="search-input"
      />
    </div>

    <!-- Role Filter -->
    <div class="role-filter">
      <span class="filter-label">Filter by role:</span>
      <button 
        @click="roleFilter = 'all'" 
        class="filter-btn" 
        :class="{ active: roleFilter === 'all' }"
      >
        All
      </button>
      <button 
        @click="roleFilter = 'tourist'" 
        class="filter-btn" 
        :class="{ active: roleFilter === 'tourist' }"
      >
        Tourists
      </button>
      <button 
        @click="roleFilter = 'guide'" 
        class="filter-btn" 
        :class="{ active: roleFilter === 'guide' }"
      >
        Guides
      </button>
    </div>

    <!-- Tabs -->
    <div class="tabs">
      <button 
        @click="activeTab = 'recommended'" 
        class="tab" 
        :class="{ active: activeTab === 'recommended' }"
      >
        Recommended
      </button>
      <button 
        @click="activeTab = 'all'" 
        class="tab" 
        :class="{ active: activeTab === 'all' }"
      >
        All Users
      </button>
    </div>

    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="displayedUsers.length === 0" class="empty">
      {{ searchQuery ? 'No users found matching your search.' : 'No users to display.' }}
    </div>
    <div v-else class="users-grid">
      <div v-for="user in displayedUsers" :key="user.id" class="user-card">
        <div class="user-header">
          <div class="user-avatar">
            <img v-if="user.profile_image" :src="user.profile_image" :alt="user.username" />
            <div v-else class="avatar-placeholder">{{ user.username.charAt(0).toUpperCase() }}</div>
          </div>
          <div class="user-info">
            <h3 @click="goToProfile(user.username)" class="username">{{ user.username }}</h3>
            <p class="name">{{ user.name }} {{ user.surname }}</p>
          </div>
        </div>

        <div class="user-details">
          <p v-if="user.biography" class="biography">{{ user.biography }}</p>
          <p v-if="user.motto" class="motto">"{{ user.motto }}"</p>
          
          <div class="user-stats">
            <span class="stat">
              <strong>{{ user.follower_count || 0 }}</strong> followers
            </span>
            <span class="stat">
              <strong>{{ user.following_count || 0 }}</strong> following
            </span>
          </div>

          <div class="user-roles" v-if="user.roles && user.roles.length > 0">
            <span v-for="role in user.roles" :key="role" class="role-badge" :class="role.toLowerCase()">{{ role }}</span>
          </div>
        </div>

        <button 
          @click="followUser(user)" 
          :disabled="user.following || user.loading"
          class="follow-btn"
          :class="{ following: user.following }"
        >
          {{ user.loading ? 'Loading...' : user.following ? 'Following' : 'Follow' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '../services/api'
import { authStore } from '../stores/authStore'

const router = useRouter()
const allUsers = ref([])
const recommendedUsers = ref([])
const loading = ref(true)
const error = ref(null)
const searchQuery = ref('')
const activeTab = ref('recommended')
const roleFilter = ref('all')

const displayedUsers = computed(() => {
  let users = activeTab.value === 'recommended' ? recommendedUsers.value : allUsers.value
  
  // Apply search filter
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase()
    users = users.filter(user => 
      user.username.toLowerCase().includes(query) ||
      user.name?.toLowerCase().includes(query) ||
      user.surname?.toLowerCase().includes(query)
    )
  }
  
  // Apply role filter
  if (roleFilter.value !== 'all') {
    users = users.filter(user => {
      const roles = user.roles || []
      return roles.some(role => role.toLowerCase() === roleFilter.value)
    })
  }
  
  return users
})

onMounted(async () => {
  await fetchData()
})

async function fetchData() {
  loading.value = true
  error.value = null
  
  try {
    const currentUserId = authStore.getUserId()
    
    // Get all users first
    const allUsersList = await api.getUsers()
    
    // Get users currently following
    const following = await api.getFollowing(currentUserId) || []
    
    // Get recommendations from backend (returns user IDs)
    let recommendedIds = []
    try {
      recommendedIds = await api.getRecommendations(currentUserId, 50) || []
    } catch (err) {
      console.warn('Failed to get recommendations:', err)
      // Continue without recommendations
    }
    
    // Filter out current user and already followed users
    const filteredUsers = allUsersList
      .filter(user => user.id !== currentUserId && (!Array.isArray(following) || !following.includes(user.id)))
      .map(user => ({
        ...user,
        follower_count: 0, // We'll load these on demand if needed
        following_count: 0,
        following: false,
        loading: false
      }))
    
    // Separate recommended and all users
    recommendedUsers.value = filteredUsers.filter(user => 
      Array.isArray(recommendedIds) && recommendedIds.includes(user.id)
    )
    
    allUsers.value = filteredUsers
    
    // Load counts for visible users (first batch)
    await loadCountsForUsers(displayedUsers.value.slice(0, 12))
    
  } catch (err) {
    console.error('Error loading users:', err)
    error.value = err.response?.data?.error || err.message || 'Failed to load users'
  } finally {
    loading.value = false
  }
}

async function loadCountsForUsers(users) {
  // Load counts in parallel for a batch of users
  await Promise.all(
    users.map(async (user) => {
      if (user.follower_count === 0 && user.following_count === 0) {
        try {
          const [followers, followingList] = await Promise.all([
            api.getFollowers(user.id),
            api.getFollowing(user.id)
          ])
          user.follower_count = followers?.length || 0
          user.following_count = followingList?.length || 0
        } catch (err) {
          console.warn(`Failed to load counts for user ${user.id}:`, err)
        }
      }
    })
  )
}

function filterUsers() {
  // Computed property handles filtering
}

async function followUser(user) {
  user.loading = true
  try {
    await api.follow(user.id)
    user.following = true
    user.follower_count++
  } catch (err) {
    alert(err.response?.data?.error || 'Failed to follow user')
  } finally {
    user.loading = false
  }
}

function goToProfile(username) {
  router.push(`/user/${username}`)
}
</script>

<style scoped>
.recommendations {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

h1 {
  margin-bottom: 10px;
  color: #2c3e50;
}

.subtitle {
  color: #666;
  font-size: 16px;
  margin-bottom: 20px;
}

.search-bar {
  margin-bottom: 20px;
}

.search-input {
  width: 100%;
  max-width: 500px;
  padding: 12px 20px;
  border: 2px solid #ddd;
  border-radius: 25px;
  font-size: 14px;
  transition: all 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: #42b983;
}

.tabs {
  display: flex;
  gap: 10px;
  margin-bottom: 30px;
  border-bottom: 2px solid #eee;
}

.tab {
  padding: 12px 24px;
  background: none;
  border: none;
  border-bottom: 3px solid transparent;
  color: #666;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  margin-bottom: -2px;
}

.tab:hover {
  color: #42b983;
}

.tab.active {
  color: #42b983;
  border-bottom-color: #42b983;
}

.loading, .error, .empty {
  text-align: center;
  padding: 40px;
  color: #666;
}

.error {
  color: #d32f2f;
}

.users-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

.user-card {
  background: white;
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 20px;
  transition: all 0.2s;
  display: flex;
  flex-direction: column;
}

.user-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  border-color: #42b983;
}

.user-header {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 15px;
}

.user-avatar {
  flex-shrink: 0;
}

.user-avatar img,
.avatar-placeholder {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  object-fit: cover;
}

.avatar-placeholder {
  background: linear-gradient(135deg, #42b983, #35a372);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  font-weight: bold;
}

.user-info {
  flex: 1;
  min-width: 0;
}

.username {
  margin: 0;
  color: #42b983;
  font-size: 18px;
  cursor: pointer;
  transition: color 0.2s;
}

.username:hover {
  color: #35a372;
  text-decoration: underline;
}

.name {
  margin: 5px 0 0 0;
  color: #666;
  font-size: 14px;
}

.user-details {
  margin-bottom: 15px;
  flex: 1;
}

.biography {
  color: #333;
  font-size: 14px;
  line-height: 1.5;
  margin: 10px 0;
}

.motto {
  color: #666;
  font-style: italic;
  font-size: 13px;
  margin: 10px 0;
}

.user-stats {
  display: flex;
  gap: 15px;
  margin: 15px 0;
  font-size: 14px;
  color: #666;
}

.stat strong {
  color: #2c3e50;
}

.user-roles {
  margin: 10px 0;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.role-badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  text-transform: capitalize;
}

.role-badge.tourist {
  background: #e3f2fd;
  color: #1976d2;
}

.role-badge.guide {
  background: #fff3e0;
  color: #f57c00;
}

.role-badge.admin {
  background: #f3e5f5;
  color: #7b1fa2;
}

.role-filter {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 20px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 8px;
}

.filter-label {
  font-weight: 500;
  color: #333;
  font-size: 14px;
}

.filter-btn {
  padding: 8px 16px;
  background: white;
  border: 2px solid #ddd;
  border-radius: 20px;
  color: #666;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.filter-btn:hover {
  border-color: #42b983;
  color: #42b983;
}

.filter-btn.active {
  background: #42b983;
  border-color: #42b983;
  color: white;
}

.follow-btn {
  width: 100%;
  padding: 10px;
  background: #42b983;
  color: white;
  border: none;
  border-radius: 5px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.follow-btn:hover:not(:disabled) {
  background: #35a372;
}

.follow-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.follow-btn.following {
  background: #f5f5f5;
  color: #666;
  border: 1px solid #ddd;
}

.follow-btn.following:hover:not(:disabled) {
  background: #e0e0e0;
}
</style>
