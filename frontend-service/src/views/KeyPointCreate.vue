<template>
  <div class="keypoint-create">
    <h1>Add Key Point to Tour</h1>
    <p class="subtitle">Click on the map to select a location</p>

    <div v-if="error" class="error-message">{{ error }}</div>
    <div v-if="success" class="success-message">Key point added successfully!</div>

    <div class="content-wrapper">
      <div class="map-container">
        <div id="map" ref="mapContainer"></div>
        <p class="map-hint">Click anywhere on the map to set the key point location</p>
      </div>

      <form @submit.prevent="handleSubmit" class="keypoint-form">
        <div class="form-group">
          <label for="name">Key Point Name *</label>
          <input 
            id="name"
            v-model="form.name" 
            type="text" 
            placeholder="e.g., City Museum, Central Park"
            required
          />
        </div>

        <div class="form-group">
          <label for="description">Description</label>
          <textarea 
            id="description"
            v-model="form.description" 
            placeholder="Describe this key point"
            rows="4"
          ></textarea>
        </div>

        <div class="form-group">
          <label for="imageUrl">Image URL</label>
          <input 
            id="imageUrl"
            v-model="form.imageUrl" 
            type="url" 
            placeholder="https://example.com/image.jpg"
          />
        </div>

        <div class="form-group">
          <label>Location *</label>
          <div class="location-display">
            <div v-if="form.latitude && form.longitude">
              <p><strong>Latitude:</strong> {{ form.latitude.toFixed(6) }}</p>
              <p><strong>Longitude:</strong> {{ form.longitude.toFixed(6) }}</p>
            </div>
            <p v-else class="no-location">Click on the map to select a location</p>
          </div>
        </div>

        <div class="form-actions">
          <button type="submit" :disabled="loading || !form.latitude" class="btn-primary">
            {{ loading ? 'Adding...' : 'Add Key Point' }}
          </button>
          <button type="button" @click="goBack" class="btn-secondary">Cancel</button>
        </div>
      </form>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../services/api'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'

export default {
  name: 'KeyPointCreate',
  setup() {
    const route = useRoute()
    const router = useRouter()
    const tourId = route.params.tourId

    const mapContainer = ref(null)
    let map = null
    let marker = null

    const form = ref({
      name: '',
      description: '',
      imageUrl: '',
      latitude: null,
      longitude: null
    })

    const loading = ref(false)
    const error = ref('')
    const success = ref(false)

    const initMap = () => {
      // Initialize map centered on Belgrade, Serbia
      map = L.map('map').setView([44.8176, 20.4633], 13)

      L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
      }).addTo(map)

      // Add click event to map
      map.on('click', (e) => {
        const { lat, lng } = e.latlng
        
        // Remove existing marker if any
        if (marker) {
          map.removeLayer(marker)
        }

        // Add new marker
        marker = L.marker([lat, lng]).addTo(map)
        
        // Update form
        form.value.latitude = lat
        form.value.longitude = lng
      })
    }

    const handleSubmit = async () => {
      if (!form.value.latitude || !form.value.longitude) {
        error.value = 'Please select a location on the map'
        return
      }

      loading.value = true
      error.value = ''
      success.value = false

      try {
        const keyPointData = {
          name: form.value.name,
          description: form.value.description,
          imageUrl: form.value.imageUrl,
          latitude: form.value.latitude,
          longitude: form.value.longitude
        }

        await api.createKeyPoint(tourId, keyPointData)
        success.value = true

        setTimeout(() => {
          router.push(`/tours/${tourId}`)
        }, 1500)
      } catch (err) {
        error.value = err.response?.data?.error || err.message || 'Failed to create key point'
      } finally {
        loading.value = false
      }
    }

    const goBack = () => {
      router.push(`/tours/${tourId}`)
    }

    onMounted(() => {
      initMap()
    })

    onUnmounted(() => {
      if (map) {
        map.remove()
      }
    })

    return {
      mapContainer,
      form,
      loading,
      error,
      success,
      handleSubmit,
      goBack
    }
  }
}
</script>

<style scoped>
.keypoint-create {
  max-width: 1400px;
  margin: 2rem auto;
  padding: 2rem;
}

h1 {
  color: #2c3e50;
  margin-bottom: 0.5rem;
}

.subtitle {
  color: #666;
  margin-bottom: 2rem;
}

.error-message {
  background: #fee;
  color: #c33;
  padding: 1rem;
  border-radius: 8px;
  margin-bottom: 1rem;
  border: 1px solid #fcc;
}

.success-message {
  background: #efe;
  color: #3c3;
  padding: 1rem;
  border-radius: 8px;
  margin-bottom: 1rem;
  border: 1px solid #cfc;
}

.content-wrapper {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
}

.map-container {
  position: relative;
}

#map {
  width: 100%;
  height: 500px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  z-index: 1;
}

.map-hint {
  margin-top: 0.5rem;
  color: #666;
  font-size: 0.9rem;
  font-style: italic;
}

.keypoint-form {
  background: white;
  padding: 2rem;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 600;
  color: #333;
}

.form-group input[type="text"],
.form-group input[type="url"],
.form-group textarea {
  width: 100%;
  padding: 0.75rem;
  border: 2px solid #ddd;
  border-radius: 8px;
  font-size: 1rem;
  transition: border-color 0.2s;
}

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #42b983;
}

.location-display {
  background: #f8f9fa;
  padding: 1rem;
  border-radius: 8px;
  border: 2px solid #ddd;
}

.location-display p {
  margin: 0.25rem 0;
  color: #333;
}

.no-location {
  color: #999;
  font-style: italic;
}

.form-actions {
  display: flex;
  gap: 1rem;
  margin-top: 2rem;
}

.btn-primary,
.btn-secondary {
  padding: 0.75rem 2rem;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #42b983;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #35a372;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background: #f5f5f5;
  color: #666;
}

.btn-secondary:hover {
  background: #e0e0e0;
}

@media (max-width: 1024px) {
  .content-wrapper {
    grid-template-columns: 1fr;
  }
}
</style>
