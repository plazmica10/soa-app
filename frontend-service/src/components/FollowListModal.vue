<template>
  <div v-if="isOpen" class="modal-overlay" @click="close">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h2>{{ title }}</h2>
        <button @click="close" class="close-btn">&times;</button>
      </div>
      
      <div v-if="loading" class="loading">Loading...</div>
      
      <div v-else-if="error" class="error-message">{{ error }}</div>
      
      <div v-else-if="users.length === 0" class="empty-message">
        No {{ type }} yet
      </div>
      
      <div v-else class="users-list">
        <div v-for="user in users" :key="user.id" class="user-item">
          <router-link :to="`/user/${user.username}`" @click="close" class="user-link">
            <div class="user-avatar">
              <img v-if="user.profile_image" :src="user.profile_image" :alt="user.username" />
              <div v-else class="avatar-placeholder">
                {{ user.name?.charAt(0) }}{{ user.surname?.charAt(0) }}
              </div>
            </div>
            <div class="user-info">
              <div class="user-name">{{ user.name }} {{ user.surname }}</div>
              <div class="user-username">@{{ user.username }}</div>
            </div>
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, watch } from 'vue'
import { api } from '../services/api'

export default {
  name: 'FollowListModal',
  props: {
    isOpen: {
      type: Boolean,
      required: true
    },
    userId: {
      type: String,
      required: true
    },
    type: {
      type: String,
      required: true,
      validator: (value) => ['followers', 'following'].includes(value)
    }
  },
  emits: ['close'],
  setup(props, { emit }) {
    const users = ref([])
    const loading = ref(false)
    const error = ref('')

    const title = ref('')

    const fetchUsers = async () => {
      if (!props.isOpen || !props.userId) return
      
      loading.value = true
      error.value = ''
      users.value = []
      
      try {
        let userIds = []
        if (props.type === 'followers') {
          title.value = 'Followers'
          userIds = await api.getFollowers(props.userId)
        } else {
          title.value = 'Following'
          userIds = await api.getFollowing(props.userId)
        }
        
        if (userIds && userIds.length > 0) {
          // Fetch user details for each ID
          const userPromises = userIds.map(id => api.getUserById(id))
          const userDetails = await Promise.all(userPromises)
          users.value = userDetails.filter(u => u != null)
        }
      } catch (err) {
        console.error('Failed to load users:', err)
        error.value = 'Failed to load users'
      } finally {
        loading.value = false
      }
    }

    watch(() => props.isOpen, (newVal) => {
      if (newVal) {
        fetchUsers()
      }
    })

    watch(() => props.type, () => {
      if (props.isOpen) {
        fetchUsers()
      }
    })

    const close = () => {
      emit('close')
    }

    return {
      users,
      loading,
      error,
      title,
      close
    }
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-content {
  background: white;
  border-radius: 16px;
  width: 100%;
  max-width: 500px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e0e0e0;
}

.modal-header h2 {
  margin: 0;
  font-size: 1.5rem;
  color: #333;
}

.close-btn {
  background: none;
  border: none;
  font-size: 2rem;
  color: #999;
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #f0f0f0;
  color: #333;
}

.loading {
  padding: 3rem;
  text-align: center;
  color: #666;
}

.error-message {
  padding: 2rem;
  text-align: center;
  color: #c33;
}

.empty-message {
  padding: 3rem;
  text-align: center;
  color: #999;
  font-style: italic;
}

.users-list {
  overflow-y: auto;
  padding: 1rem;
}

.user-item {
  margin-bottom: 0.5rem;
}

.user-link {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem;
  border-radius: 8px;
  text-decoration: none;
  color: inherit;
  transition: background 0.2s;
}

.user-link:hover {
  background: #f5f5f5;
}

.user-avatar {
  flex-shrink: 0;
}

.user-avatar img {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  object-fit: cover;
}

.avatar-placeholder {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 1.2rem;
}

.user-info {
  flex: 1;
}

.user-name {
  font-weight: 600;
  color: #333;
  margin-bottom: 0.25rem;
}

.user-username {
  color: #666;
  font-size: 0.9rem;
}
</style>
