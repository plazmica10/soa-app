<template>
  <div class="user-profile-container">
    <div v-if="loading" class="loading">Loading profile...</div>
    
    <div v-else-if="error" class="error-message">
      {{ error }}
    </div>

    <div v-else class="profile-card">
      <div class="profile-header">
        <div class="profile-image-container">
          <img 
            v-if="user.profile_image" 
            :src="user.profile_image" 
            :alt="user.username"
            class="profile-image"
          />
          <div v-else class="profile-image-placeholder">
            {{ user.name?.charAt(0) }}{{ user.surname?.charAt(0) }}
          </div>
        </div>
        <div class="profile-info">
          <h1>{{ user.name }} {{ user.surname }}</h1>
          <p class="username">@{{ user.username }}</p>
          <div class="roles">
            <span class="role-badge" v-for="role in user.roles" :key="role">
              {{ role }}
            </span>
          </div>
          <div class="stats">
            <div class="stat-item" @click="openFollowersModal">
              <span class="stat-value">{{ followersCount }}</span>
              <span class="stat-label">Followers</span>
            </div>
            <div class="stat-item" @click="openFollowingModal">
              <span class="stat-value">{{ followingCount }}</span>
              <span class="stat-label">Following</span>
            </div>
          </div>
        </div>
        <div v-if="isAuthenticated && !isOwnProfile" class="follow-actions">
          <button 
            v-if="!isFollowing"
            @click="handleFollow"
            class="follow-btn"
            :disabled="followLoading"
          >
            {{ followLoading ? 'Following...' : 'Follow' }}
          </button>
          <button 
            v-else
            @click="handleUnfollow"
            class="unfollow-btn"
            :disabled="followLoading"
          >
            {{ followLoading ? 'Unfollowing...' : 'Unfollow' }}
          </button>
        </div>
      </div>

      <div class="profile-body">
        <div class="profile-section" v-if="user.motto">
          <h3>Motto</h3>
          <p class="motto">"{{ user.motto }}"</p>
        </div>

        <div class="profile-section" v-if="user.biography">
          <h3>Biography</h3>
          <p class="biography">{{ user.biography }}</p>
        </div>

        <div class="profile-section">
          <h3>Contact Information</h3>
          <div class="info-grid">
            <div class="info-item">
              <span class="info-label">Email:</span>
              <span class="info-value">{{ user.email }}</span>
            </div>
          </div>
        </div>

        <div class="profile-section" v-if="hasAddress">
          <h3>Address</h3>
          <div class="info-grid">
            <div class="info-item" v-if="user.address.street">
              <span class="info-label">Street:</span>
              <span class="info-value">{{ user.address.street }}</span>
            </div>
            <div class="info-item" v-if="user.address.city">
              <span class="info-label">City:</span>
              <span class="info-value">{{ user.address.city }}</span>
            </div>
            <div class="info-item" v-if="user.address.state">
              <span class="info-label">State:</span>
              <span class="info-value">{{ user.address.state }}</span>
            </div>
            <div class="info-item" v-if="user.address.postal_code">
              <span class="info-label">Postal Code:</span>
              <span class="info-value">{{ user.address.postal_code }}</span>
            </div>
            <div class="info-item" v-if="user.address.country">
              <span class="info-label">Country:</span>
              <span class="info-value">{{ user.address.country }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <FollowListModal
      :isOpen="isModalOpen"
      :userId="user.id"
      :type="modalType"
      @close="closeModal"
    />
  </div>
</template>

<script>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '../services/api'
import { authStore } from '../stores/authStore'
import { authService } from '../services/auth'
import FollowListModal from '../components/FollowListModal.vue'

export default {
  name: 'UserProfile',
  components: {
    FollowListModal
  },
  setup() {
    const route = useRoute()
    const user = ref({})
    const loading = ref(true)
    const error = ref('')
    const followLoading = ref(false)
    const isFollowing = ref(false)
    const followersCount = ref(0)
    const followingCount = ref(0)
    const isModalOpen = ref(false)
    const modalType = ref('followers')

    const isAuthenticated = computed(() => authStore.isAuthenticated.value)
    
    const isOwnProfile = computed(() => {
      const currentUser = authService.getUserFromToken()
      return currentUser && user.value.username === authStore.username.value
    })

    const hasAddress = computed(() => {
      const addr = user.value.address
      return addr && (addr.street || addr.city || addr.state || addr.postal_code || addr.country)
    })

    const fetchProfile = async () => {
      loading.value = true
      error.value = ''
      try {
        const username = route.params.username
        const users = await api.getUserByUsername(username)
        
        if (users && users.length > 0) {
          user.value = users[0]
          await loadFollowStats()
        } else {
          error.value = 'User not found'
        }
      } catch (err) {
        error.value = 'Failed to load profile. ' + (err.message || '')
      } finally {
        loading.value = false
      }
    }

    const loadFollowStats = async () => {
      if (!user.value.id) return
      
      try {
        const [followers, following] = await Promise.all([
          api.getFollowers(user.value.id),
          api.getFollowing(user.value.id)
        ])
        
        followersCount.value = followers?.length || 0
        followingCount.value = following?.length || 0

        // Check if current user is following this user
        if (isAuthenticated.value && !isOwnProfile.value) {
          const currentUserId = authStore.getUserId()
          if (currentUserId) {
            isFollowing.value = await api.isFollowing(currentUserId, user.value.id)
          }
        }
      } catch (err) {
        console.error('Failed to load follow stats:', err)
      }
    }

    const handleFollow = async () => {
      followLoading.value = true
      try {
        await api.follow(user.value.id)
        isFollowing.value = true
        followersCount.value++
      } catch (err) {
        error.value = 'Failed to follow user. ' + (err.message || '')
      } finally {
        followLoading.value = false
      }
    }

    const handleUnfollow = async () => {
      followLoading.value = true
      try {
        await api.unfollow(user.value.id)
        isFollowing.value = false
        followersCount.value--
      } catch (err) {
        error.value = 'Failed to unfollow user. ' + (err.message || '')
      } finally {
        followLoading.value = false
      }
    }

    const openFollowersModal = () => {
      modalType.value = 'followers'
      isModalOpen.value = true
    }

    const openFollowingModal = () => {
      modalType.value = 'following'
      isModalOpen.value = true
    }

    const closeModal = () => {
      isModalOpen.value = false
    }

    watch(() => route.params.username, () => {
      if (route.params.username) {
        fetchProfile()
      }
    })

    onMounted(() => {
      fetchProfile()
    })

    return {
      user,
      loading,
      error,
      hasAddress,
      isAuthenticated,
      isOwnProfile,
      isFollowing,
      followLoading,
      followersCount,
      followingCount,
      handleFollow,
      handleUnfollow,
      isModalOpen,
      modalType,
      openFollowersModal,
      openFollowingModal,
      closeModal
    }
  }
}
</script>

<style scoped>
.user-profile-container {
  max-width: 900px;
  margin: 2rem auto;
  padding: 2rem;
}

.loading {
  text-align: center;
  padding: 3rem;
  color: #666;
  font-size: 1.2rem;
}

.error-message {
  background: #fee;
  color: #c33;
  padding: 1rem;
  border-radius: 8px;
  border: 1px solid #fcc;
}

.profile-card {
  background: white;
  border-radius: 16px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.profile-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 3rem 2rem;
  color: white;
  position: relative;
  display: flex;
  align-items: flex-start;
  gap: 2rem;
}

.profile-image-container {
  flex-shrink: 0;
}

.profile-image {
  width: 120px;
  height: 120px;
  border-radius: 50%;
  object-fit: cover;
  border: 4px solid white;
}

.profile-image-placeholder {
  width: 120px;
  height: 120px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 3rem;
  font-weight: 600;
  border: 4px solid white;
}

.profile-info {
  flex: 1;
}

.profile-info h1 {
  margin: 0 0 0.5rem 0;
  font-size: 2rem;
}

.username {
  margin: 0 0 1rem 0;
  opacity: 0.9;
  font-size: 1.1rem;
}

.roles {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.role-badge {
  padding: 0.25rem 0.75rem;
  background: rgba(255, 255, 255, 0.3);
  border-radius: 12px;
  font-size: 0.85rem;
  text-transform: capitalize;
}

.stats {
  display: flex;
  gap: 2rem;
  margin-top: 1rem;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  cursor: pointer;
  transition: all 0.2s;
  padding: 0.5rem;
  border-radius: 8px;
}

.stat-item:hover {
  background: rgba(255, 255, 255, 0.2);
  transform: translateY(-2px);
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 600;
}

.stat-label {
  font-size: 0.9rem;
  opacity: 0.9;
}

.follow-actions {
  flex-shrink: 0;
}

.follow-btn,
.unfollow-btn {
  padding: 0.75rem 2rem;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
}

.follow-btn {
  background: white;
  color: #667eea;
}

.follow-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.unfollow-btn {
  background: rgba(255, 255, 255, 0.2);
  color: white;
  border: 2px solid white;
}

.unfollow-btn:hover:not(:disabled) {
  background: #ff4757;
  border-color: #ff4757;
}

.follow-btn:disabled,
.unfollow-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.profile-body {
  padding: 2rem;
}

.profile-section {
  margin-bottom: 2rem;
}

.profile-section:last-child {
  margin-bottom: 0;
}

.profile-section h3 {
  color: #333;
  margin: 0 0 1rem 0;
  font-size: 1.3rem;
  border-bottom: 2px solid #667eea;
  padding-bottom: 0.5rem;
}

.motto {
  font-style: italic;
  font-size: 1.2rem;
  color: #555;
  margin: 0;
  padding: 1rem;
  background: #f8f9fa;
  border-left: 4px solid #667eea;
  border-radius: 4px;
}

.biography {
  line-height: 1.6;
  color: #555;
  margin: 0;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1rem;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.info-label {
  font-weight: 600;
  color: #666;
  font-size: 0.9rem;
}

.info-value {
  color: #333;
  font-size: 1rem;
}

@media (max-width: 768px) {
  .profile-header {
    flex-direction: column;
    align-items: center;
    text-align: center;
  }

  .follow-actions {
    width: 100%;
  }

  .follow-btn,
  .unfollow-btn {
    width: 100%;
  }
}
</style>
