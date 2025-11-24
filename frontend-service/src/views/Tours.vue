<template>
  <div class="tours-container">
    <!-- Guide View -->
    <div v-if="isGuide">
      <h1>Tours</h1>
      <p class="subtitle">Manage and create amazing tours</p>

      <div class="actions">
        <router-link to="/tours/my-tours" class="action-card">
          <div class="icon">📝</div>
          <h2>My Tours</h2>
          <p>View and manage your tours</p>
        </router-link>

        <router-link to="/tours/create" class="action-card">
          <div class="icon">➕</div>
          <h2>Create Tour</h2>
          <p>Start creating a new tour</p>
        </router-link>
      </div>
    </div>

    <!-- Tourist View -->
    <div v-else>
      <h1>Discover Tours</h1>
      <p class="subtitle">Explore tours from guides you follow</p>
      <ToursList />
    </div>
  </div>
</template>

<script>
import { computed } from 'vue'
import { authStore } from '../stores/authStore'
import ToursList from './ToursList.vue'

export default {
  name: 'Tours',
  components: {
    ToursList
  },
  setup() {
    const isGuide = computed(() => authStore.isGuide())

    return {
      isGuide
    }
  }
}
</script>

<style scoped>
.tours-container {
  max-width: 1200px;
  margin: 2rem auto;
  padding: 2rem;
  text-align: center;
}

h1 {
  font-size: 3rem;
  color: #2c3e50;
  margin-bottom: 0.5rem;
}

.subtitle {
  font-size: 1.2rem;
  color: #666;
  margin-bottom: 3rem;
}

.actions {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
  max-width: 800px;
  margin: 0 auto;
}

.action-card {
  background: white;
  padding: 3rem 2rem;
  border-radius: 16px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
  text-decoration: none;
  transition: all 0.3s;
  border: 2px solid transparent;
}

.action-card:hover {
  transform: translateY(-8px);
  box-shadow: 0 8px 30px rgba(66, 185, 131, 0.3);
  border-color: #42b983;
}

.icon {
  font-size: 4rem;
  margin-bottom: 1rem;
}

.action-card h2 {
  color: #2c3e50;
  margin-bottom: 0.5rem;
}

.action-card p {
  color: #666;
  margin: 0;
}

@media (max-width: 768px) {
  h1 {
    font-size: 2rem;
  }

  .actions {
    grid-template-columns: 1fr;
  }
}
</style>
