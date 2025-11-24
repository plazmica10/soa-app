<template>
  <div class="admin-container">
    <div class="admin-header">
      <h1>User Management</h1>
      <p class="subtitle">Admin Panel</p>
    </div>

    <div v-if="error" class="error-message">
      {{ error }}
    </div>

    <div v-if="loading" class="loading">Loading users...</div>

    <div v-else class="users-table-container">
      <table class="users-table">
        <thead>
          <tr>
            <th>Username</th>
            <th>Name</th>
            <th>Email</th>
            <th>Roles</th>
            <th>Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in users" :key="user.id">
            <td>
              <router-link :to="`/user/${user.username}`" class="username-link">
                {{ user.username }}
              </router-link>
            </td>
            <td>{{ user.name }} {{ user.surname }}</td>
            <td>{{ user.email }}</td>
            <td>
              <span class="role-badge" v-for="role in user.roles" :key="role">
                {{ role }}
              </span>
            </td>
            <td>
              <span class="status-badge" :class="user.is_blocked ? 'blocked' : 'active'">
                {{ user.is_blocked ? 'Blocked' : 'Active' }}
              </span>
            </td>
            <td>
              <button 
                v-if="!user.is_blocked && !user.roles.includes('admin')"
                @click="blockUser(user.id)"
                class="action-btn block-btn"
                :disabled="actionLoading[user.id]"
              >
                {{ actionLoading[user.id] ? 'Blocking...' : 'Block' }}
              </button>
              <button 
                v-if="user.is_blocked"
                @click="unblockUser(user.id)"
                class="action-btn unblock-btn"
                :disabled="actionLoading[user.id]"
              >
                {{ actionLoading[user.id] ? 'Unblocking...' : 'Unblock' }}
              </button>
              <span v-if="user.roles.includes('admin')" class="admin-label">Admin</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, reactive } from 'vue'
import { api } from '../services/api'

export default {
  name: 'AdminUsers',
  setup() {
    const users = ref([])
    const loading = ref(true)
    const error = ref('')
    const actionLoading = reactive({})

    const fetchUsers = async () => {
      loading.value = true
      error.value = ''
      try {
        users.value = await api.getUsers()
      } catch (err) {
        error.value = 'Failed to load users. ' + (err.message || '')
      } finally {
        loading.value = false
      }
    }

    const blockUser = async (userId) => {
      actionLoading[userId] = true
      try {
        await api.blockUser(userId)
        await fetchUsers()
      } catch (err) {
        error.value = 'Failed to block user. ' + (err.message || '')
      } finally {
        actionLoading[userId] = false
      }
    }

    const unblockUser = async (userId) => {
      actionLoading[userId] = true
      try {
        await api.unblockUser(userId)
        await fetchUsers()
      } catch (err) {
        error.value = 'Failed to unblock user. ' + (err.message || '')
      } finally {
        actionLoading[userId] = false
      }
    }

    onMounted(() => {
      fetchUsers()
    })

    return {
      users,
      loading,
      error,
      actionLoading,
      blockUser,
      unblockUser
    }
  }
}
</script>

<style scoped>
.admin-container {
  max-width: 1200px;
  margin: 2rem auto;
  padding: 2rem;
}

.admin-header {
  margin-bottom: 2rem;
}

.admin-header h1 {
  font-size: 2.5rem;
  color: #333;
  margin-bottom: 0.5rem;
}

.subtitle {
  color: #666;
  font-size: 1.1rem;
}

.error-message {
  background: #fee;
  color: #c33;
  padding: 1rem;
  border-radius: 8px;
  border: 1px solid #fcc;
  margin-bottom: 1rem;
}

.loading {
  text-align: center;
  padding: 3rem;
  color: #666;
  font-size: 1.2rem;
}

.users-table-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.users-table {
  width: 100%;
  border-collapse: collapse;
}

.users-table thead {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.users-table th {
  padding: 1rem;
  text-align: left;
  font-weight: 600;
}

.users-table td {
  padding: 1rem;
  border-bottom: 1px solid #e0e0e0;
}

.users-table tbody tr:hover {
  background: #f8f9fa;
}

.role-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  background: #667eea;
  color: white;
  border-radius: 12px;
  font-size: 0.85rem;
  margin-right: 0.5rem;
}

.status-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.85rem;
  font-weight: 500;
}

.status-badge.active {
  background: #d4edda;
  color: #155724;
}

.status-badge.blocked {
  background: #f8d7da;
  color: #721c24;
}

.action-btn {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 6px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
  margin-right: 0.5rem;
}

.block-btn {
  background: #ff4757;
  color: white;
}

.block-btn:hover:not(:disabled) {
  background: #ee3344;
}

.unblock-btn {
  background: #4CAF50;
  color: white;
}

.unblock-btn:hover:not(:disabled) {
  background: #45a049;
}

.action-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.admin-label {
  color: #666;
  font-style: italic;
}

.username-link {
  color: #667eea;
  text-decoration: none;
  font-weight: 500;
  transition: color 0.3s;
}

.username-link:hover {
  color: #764ba2;
  text-decoration: underline;
}

@media (max-width: 768px) {
  .users-table-container {
    overflow-x: auto;
  }

  .users-table {
    min-width: 800px;
  }
}
</style>
