<template>
  <div class="edit-profile-container">
    <div class="edit-card">
      <h1>Edit Profile</h1>
      <p class="subtitle">Update your information</p>

      <form @submit.prevent="handleUpdate" class="edit-form">
        <div v-if="error" class="error-message">
          {{ error }}
        </div>

        <div v-if="success" class="success-message">
          Profile updated successfully!
        </div>

        <div class="form-row">
          <div class="form-group">
            <label for="name">First Name *</label>
            <input
              type="text"
              id="name"
              v-model="formData.name"
              required
            />
          </div>

          <div class="form-group">
            <label for="surname">Last Name *</label>
            <input
              type="text"
              id="surname"
              v-model="formData.surname"
              required
            />
          </div>
        </div>

        <div class="form-group">
          <label for="email">Email *</label>
          <input
            type="email"
            id="email"
            v-model="formData.email"
            required
          />
        </div>

        <div class="form-group">
          <label for="profile_image">Profile Image URL</label>
          <input
            type="url"
            id="profile_image"
            v-model="formData.profile_image"
            placeholder="https://example.com/image.jpg"
          />
        </div>

        <div class="form-group">
          <label for="motto">Motto / Quote</label>
          <input
            type="text"
            id="motto"
            v-model="formData.motto"
            placeholder="Your personal motto..."
            maxlength="200"
          />
        </div>

        <div class="form-group">
          <label for="biography">Biography</label>
          <textarea
            id="biography"
            v-model="formData.biography"
            placeholder="Tell us about yourself..."
            rows="5"
            maxlength="1000"
          ></textarea>
        </div>

        <div class="address-section">
          <h3>Address Information</h3>
          
          <div class="form-group">
            <label for="street">Street</label>
            <input
              type="text"
              id="street"
              v-model="formData.address.street"
            />
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="city">City</label>
              <input
                type="text"
                id="city"
                v-model="formData.address.city"
              />
            </div>

            <div class="form-group">
              <label for="state">State</label>
              <input
                type="text"
                id="state"
                v-model="formData.address.state"
              />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="postalCode">Postal Code</label>
              <input
                type="text"
                id="postalCode"
                v-model="formData.address.postal_code"
              />
            </div>

            <div class="form-group">
              <label for="country">Country</label>
              <input
                type="text"
                id="country"
                v-model="formData.address.country"
              />
            </div>
          </div>
        </div>

        <div class="form-actions">
          <button type="submit" class="submit-btn" :disabled="loading">
            {{ loading ? 'Saving...' : 'Save Changes' }}
          </button>
          <router-link to="/profile" class="cancel-btn">
            Cancel
          </router-link>
        </div>
      </form>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../services/api'

export default {
  name: 'EditProfile',
  setup() {
    const router = useRouter()
    const loading = ref(false)
    const error = ref('')
    const success = ref(false)

    const formData = ref({
      name: '',
      surname: '',
      email: '',
      profile_image: '',
      biography: '',
      motto: '',
      address: {
        street: '',
        city: '',
        state: '',
        postal_code: '',
        country: ''
      }
    })

    const loadProfile = async () => {
      try {
        const user = await api.getCurrentUser()
        formData.value = {
          name: user.name || '',
          surname: user.surname || '',
          email: user.email || '',
          profile_image: user.profile_image || '',
          biography: user.biography || '',
          motto: user.motto || '',
          address: {
            street: user.address?.street || '',
            city: user.address?.city || '',
            state: user.address?.state || '',
            postal_code: user.address?.postal_code || '',
            country: user.address?.country || ''
          }
        }
      } catch (err) {
        error.value = 'Failed to load profile data'
      }
    }

    const handleUpdate = async () => {
      loading.value = true
      error.value = ''
      success.value = false

      try {
        await api.updateCurrentUser(formData.value)
        success.value = true
        setTimeout(() => {
          router.push('/profile')
        }, 1500)
      } catch (err) {
        error.value = err.message || 'Failed to update profile'
      } finally {
        loading.value = false
      }
    }

    onMounted(() => {
      loadProfile()
    })

    return {
      formData,
      loading,
      error,
      success,
      handleUpdate
    }
  }
}
</script>

<style scoped>
.edit-profile-container {
  min-height: calc(100vh - 80px);
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 2rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.edit-card {
  background: white;
  padding: 3rem;
  border-radius: 16px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1);
  width: 100%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
}

.edit-card h1 {
  margin: 0 0 0.5rem 0;
  color: #333;
  font-size: 2rem;
  text-align: center;
}

.subtitle {
  margin: 0 0 2rem 0;
  color: #666;
  text-align: center;
}

.edit-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.error-message {
  background: #fee;
  color: #c33;
  padding: 1rem;
  border-radius: 8px;
  border: 1px solid #fcc;
  font-size: 0.9rem;
}

.success-message {
  background: #d4edda;
  color: #155724;
  padding: 1rem;
  border-radius: 8px;
  border: 1px solid #c3e6cb;
  font-size: 0.9rem;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group label {
  font-weight: 500;
  color: #333;
  font-size: 0.9rem;
}

.form-group input,
.form-group textarea {
  padding: 0.875rem;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 1rem;
  transition: border-color 0.3s;
  font-family: inherit;
}

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #667eea;
}

.address-section {
  background: #f8f9fa;
  padding: 1.5rem;
  border-radius: 12px;
  margin-top: 0.5rem;
}

.address-section h3 {
  margin: 0 0 1.25rem 0;
  color: #333;
  font-size: 1.1rem;
  border-bottom: 2px solid #667eea;
  padding-bottom: 0.5rem;
}

.form-actions {
  display: flex;
  gap: 1rem;
  margin-top: 0.5rem;
}

.submit-btn {
  flex: 1;
  padding: 1rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 1.05rem;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.3s, box-shadow 0.3s;
}

.submit-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.4);
}

.submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.cancel-btn {
  flex: 1;
  padding: 1rem;
  background: #e0e0e0;
  color: #333;
  text-decoration: none;
  border-radius: 8px;
  font-size: 1.05rem;
  font-weight: 600;
  text-align: center;
  transition: background 0.3s;
}

.cancel-btn:hover {
  background: #d0d0d0;
}

@media (max-width: 768px) {
  .form-row {
    grid-template-columns: 1fr;
  }

  .edit-card {
    padding: 2rem;
  }

  .form-actions {
    flex-direction: column;
  }
}
</style>
