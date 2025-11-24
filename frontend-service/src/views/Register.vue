<template>
  <div class="auth-container">
    <div class="auth-card register-card">
      <h1>Create Account</h1>
      <p class="subtitle">Join us today</p>

      <form @submit.prevent="handleRegister" class="auth-form">
        <div v-if="error" class="error-message">
          {{ error }}
        </div>

        <div class="form-row">
          <div class="form-group">
            <label for="name">First Name *</label>
            <input
              type="text"
              id="name"
              v-model="formData.name"
              placeholder="John"
              required
            />
          </div>

          <div class="form-group">
            <label for="surname">Last Name *</label>
            <input
              type="text"
              id="surname"
              v-model="formData.surname"
              placeholder="Doe"
              required
            />
          </div>
        </div>

        <div class="form-group">
          <label for="username">Username *</label>
          <input
            type="text"
            id="username"
            v-model="formData.username"
            placeholder="johndoe"
            required
          />
        </div>

        <div class="form-group">
          <label for="email">Email *</label>
          <input
            type="email"
            id="email"
            v-model="formData.email"
            placeholder="john@example.com"
            required
          />
        </div>

        <div class="form-group">
          <label for="role">Role *</label>
          <select
            id="role"
            v-model="selectedRole"
            required
            class="role-select"
          >
            <option value="">Select your role</option>
            <option value="tourist">Tourist</option>
            <option value="guide">Guide</option>
          </select>
        </div>

        <div class="form-group">
          <label for="password">Password *</label>
          <input
            type="password"
            id="password"
            v-model="formData.password"
            placeholder="Minimum 6 characters"
            required
            minlength="6"
          />
        </div>

        <div class="address-section">
          <h3>Address Information</h3>
          
          <div class="form-group">
            <label for="street">Street</label>
            <input
              type="text"
              id="street"
              v-model="formData.address.street"
              placeholder="123 Main St"
            />
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="city">City</label>
              <input
                type="text"
                id="city"
                v-model="formData.address.city"
                placeholder="New York"
              />
            </div>

            <div class="form-group">
              <label for="state">State</label>
              <input
                type="text"
                id="state"
                v-model="formData.address.state"
                placeholder="NY"
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
                placeholder="10001"
              />
            </div>

            <div class="form-group">
              <label for="country">Country</label>
              <input
                type="text"
                id="country"
                v-model="formData.address.country"
                placeholder="USA"
              />
            </div>
          </div>

          <div class="location-group">
            <p class="location-label">Location Coordinates</p>
            <button 
              type="button" 
              @click="geocodeAddress" 
              class="geocode-btn"
              :disabled="geocoding || !canGeocode"
            >
              {{ geocoding ? 'Getting Coordinates...' : 'Get Coordinates from Address' }}
            </button>
            <p v-if="geocodeError" class="geocode-error">{{ geocodeError }}</p>
            <p v-if="geocodeSuccess" class="geocode-success">✓ Coordinates obtained successfully</p>
            <div class="form-row" v-if="coordinates.longitude !== null">
              <div class="form-group">
                <label for="longitude">Longitude</label>
                <input
                  type="number"
                  id="longitude"
                  v-model.number="coordinates.longitude"
                  placeholder="-74.0060"
                  step="any"
                  readonly
                />
              </div>

              <div class="form-group">
                <label for="latitude">Latitude</label>
                <input
                  type="number"
                  id="latitude"
                  v-model.number="coordinates.latitude"
                  placeholder="40.7128"
                  step="any"
                  readonly
                />
              </div>
            </div>
          </div>
        </div>

        <button type="submit" class="submit-btn" :disabled="loading">
          {{ loading ? 'Creating Account...' : 'Register' }}
        </button>

        <p class="switch-auth">
          Already have an account? 
          <router-link to="/login">Login here</router-link>
        </p>
      </form>
    </div>
  </div>
</template>

<script>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { authService } from '../services/auth'

export default {
  name: 'Register',
  setup() {
    const router = useRouter()
    const loading = ref(false)
    const error = ref('')

    const formData = ref({
      username: '',
      password: '',
      email: '',
      name: '',
      surname: '',
      roles: ['user'],
      address: {
        street: '',
        city: '',
        state: '',
        postal_code: '',
        country: ''
      },
      profile_image: '',
      biography: '',
      motto: ''
    })

    const selectedRole = ref('')

    const coordinates = ref({
      longitude: null,
      latitude: null
    })

    const geocoding = ref(false)
    const geocodeError = ref('')
    const geocodeSuccess = ref(false)

    const canGeocode = computed(() => {
      const addr = formData.value.address
      return addr.street || addr.city || addr.country
    })

    const geocodeAddress = async () => {
      geocoding.value = true
      geocodeError.value = ''
      geocodeSuccess.value = false

      try {
        const addr = formData.value.address
        const addressParts = [
          addr.street,
          addr.city,
          addr.state,
          addr.postal_code,
          addr.country
        ].filter(Boolean).join(', ')

        if (!addressParts) {
          geocodeError.value = 'Please fill in at least one address field'
          return
        }

        // Using Nominatim (OpenStreetMap) geocoding service
        const response = await fetch(
          `https://nominatim.openstreetmap.org/search?format=json&q=${encodeURIComponent(addressParts)}&limit=1`,
          {
            headers: {
              'User-Agent': 'MyApp/1.0' // Required by Nominatim
            }
          }
        )

        const data = await response.json()

        if (data && data.length > 0) {
          coordinates.value.latitude = parseFloat(data[0].lat)
          coordinates.value.longitude = parseFloat(data[0].lon)
          geocodeSuccess.value = true
        } else {
          geocodeError.value = 'Address not found. Please check your address and try again.'
        }
      } catch (err) {
        geocodeError.value = 'Failed to get coordinates. Please try again.'
        console.error('Geocoding error:', err)
      } finally {
        geocoding.value = false
      }
    }

    const handleRegister = async () => {
      loading.value = true
      error.value = ''

      try {
        // Set role based on selection
        if (selectedRole.value) {
          formData.value.roles = [selectedRole.value]
        }

        // Add location if coordinates are provided
        if (coordinates.value.longitude !== null && coordinates.value.latitude !== null) {
          formData.value.address.location = {
            type: 'Point',
            coordinates: [coordinates.value.longitude, coordinates.value.latitude]
          }
        }

        await authService.register(formData.value)
        router.push('/login')
      } catch (err) {
        // Display detailed error message from server
        error.value = err.message || 'Registration failed. Please try again.'
      } finally {
        loading.value = false
      }
    }

    return {
      formData,
      selectedRole,
      coordinates,
      loading,
      error,
      geocoding,
      geocodeError,
      geocodeSuccess,
      canGeocode,
      handleRegister,
      geocodeAddress
    }
  }
}
</script>

<style scoped>
.auth-container {
  min-height: calc(100vh - 80px);
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 2rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.register-card {
  max-width: 600px;
}

.auth-card {
  background: white;
  padding: 3rem;
  border-radius: 16px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1);
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
}

.auth-card h1 {
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

.auth-form {
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

.form-group input {
  padding: 0.875rem;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 1rem;
  transition: border-color 0.3s;
}

.form-group select.role-select {
  padding: 0.875rem;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 1rem;
  transition: border-color 0.3s;
  background: white;
  cursor: pointer;
}

.form-group input:focus,
.form-group select:focus {
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

.location-group {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #e0e0e0;
}

.location-label {
  margin: 0 0 0.75rem 0;
  color: #666;
  font-size: 0.9rem;
  font-weight: 500;
}

.geocode-btn {
  width: 100%;
  padding: 0.875rem;
  background: #4CAF50;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 0.95rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.3s;
  margin-bottom: 1rem;
}

.geocode-btn:hover:not(:disabled) {
  background: #45a049;
}

.geocode-btn:disabled {
  background: #ccc;
  cursor: not-allowed;
}

.geocode-error {
  color: #c33;
  font-size: 0.9rem;
  margin: 0.5rem 0;
}

.geocode-success {
  color: #4CAF50;
  font-size: 0.9rem;
  margin: 0.5rem 0;
  font-weight: 500;
}

.form-group input[readonly] {
  background: #f5f5f5;
  color: #666;
}

.submit-btn {
  padding: 1rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 1.05rem;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.3s, box-shadow 0.3s;
  margin-top: 0.5rem;
}

.submit-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.4);
}

.submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.switch-auth {
  text-align: center;
  color: #666;
  margin-top: 0.5rem;
}

.switch-auth a {
  color: #667eea;
  text-decoration: none;
  font-weight: 500;
}

.switch-auth a:hover {
  text-decoration: underline;
}

@media (max-width: 768px) {
  .form-row {
    grid-template-columns: 1fr;
  }

  .auth-card {
    padding: 2rem;
  }
}
</style>
