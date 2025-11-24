<template>
  <div class="tour-detail">
    <div v-if="loading" class="loading">Loading tour...</div>
    <div v-else-if="error" class="error-message">{{ error }}</div>

    <div v-else class="tour-content">
      <!-- Tour Header -->
      <div class="tour-header">
        <div>
          <h1>{{ tour.name }}</h1>
          <div class="tour-meta">
            <span class="status-badge" :class="tour.status">{{ tour.status }}</span>
            <span class="difficulty">Difficulty: {{ tour.difficulty }}</span>
            <span class="price">Price: ${{ tour.price }}</span>
          </div>
        </div>
      </div>

      <!-- Tour Description -->
      <div v-if="tour.description" class="tour-description">
        <h2>Description</h2>
        <p>{{ tour.description }}</p>
      </div>

      <!-- Tags -->
      <div v-if="tour.tags && tour.tags.length > 0" class="tour-tags">
        <h3>Tags</h3>
        <div class="tags-list">
          <span v-for="tag in tour.tags" :key="tag" class="tag">{{ tag }}</span>
        </div>
      </div>

      <!-- Key Points Section -->
      <div class="section keypoints-section">
        <div class="section-header">
          <h2>Key Points</h2>
          <router-link 
            v-if="isAuthor" 
            :to="`/tours/${tour.id}/keypoints/create`" 
            class="btn-add"
          >
            + Add Key Point
          </router-link>
        </div>

        <div v-if="loadingKeyPoints" class="loading-small">Loading key points...</div>
        <div v-else-if="keyPoints.length === 0" class="empty-message">
          No key points added yet.
        </div>
        <div v-else class="keypoints-grid">
          <div v-for="kp in keyPoints" :key="kp.id" class="keypoint-card">
            <img v-if="kp.imageUrl" :src="kp.imageUrl" :alt="kp.name" class="keypoint-image" />
            <div class="keypoint-content">
              <h3>{{ kp.name }}</h3>
              <p v-if="kp.description">{{ kp.description }}</p>
              <div class="keypoint-location">
                📍 {{ kp.latitude.toFixed(4) }}, {{ kp.longitude.toFixed(4) }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Reviews Section -->
      <div class="section reviews-section">
        <div class="section-header">
          <h2>Reviews</h2>
          <button 
            v-if="canAddReview" 
            @click="showReviewForm = true" 
            class="btn-add"
          >
            + Add Review
          </button>
        </div>

        <!-- Review Form -->
        <div v-if="showReviewForm" class="review-form">
          <h3>Write a Review</h3>
          <form @submit.prevent="submitReview">
            <div class="form-group">
              <label>Rating *</label>
              <div class="rating-input">
                <button
                  v-for="n in 5"
                  :key="n"
                  type="button"
                  @click="reviewForm.rating = n"
                  class="star-btn"
                  :class="{ active: n <= reviewForm.rating }"
                >
                  ★
                </button>
              </div>
            </div>

            <div class="form-group">
              <label>Comment</label>
              <textarea 
                v-model="reviewForm.comment" 
                placeholder="Share your experience..."
                rows="4"
              ></textarea>
            </div>

            <div class="form-group">
              <label>When did you visit?</label>
              <input 
                v-model="reviewForm.visitedAt" 
                type="date" 
                :max="today"
              />
            </div>

            <div class="form-group">
              <label>Image URLs (one per line)</label>
              <textarea 
                v-model="imagesText" 
                placeholder="https://example.com/image1.jpg&#10;https://example.com/image2.jpg"
                rows="3"
              ></textarea>
            </div>

            <div class="form-actions">
              <button type="submit" :disabled="!reviewForm.rating || reviewLoading" class="btn-primary">
                {{ reviewLoading ? 'Submitting...' : 'Submit Review' }}
              </button>
              <button type="button" @click="cancelReview" class="btn-secondary">Cancel</button>
            </div>
          </form>
        </div>

        <!-- Reviews List -->
        <div v-if="loadingReviews" class="loading-small">Loading reviews...</div>
        <div v-else-if="reviews.length === 0" class="empty-message">
          No reviews yet. Be the first to review!
        </div>
        <div v-else class="reviews-list">
          <div v-for="review in reviews" :key="review.id" class="review-card">
            <div class="review-header">
              <div>
                <h4>{{ review.authorName || 'Anonymous' }}</h4>
                <div class="review-meta">
                  <span class="rating">{{ '★'.repeat(review.rating) }}{{ '☆'.repeat(5 - review.rating) }}</span>
                  <span v-if="review.visitedAt" class="visited-date">
                    Visited: {{ formatDate(review.visitedAt) }}
                  </span>
                </div>
              </div>
              <span class="review-date">{{ formatDate(review.createdAt) }}</span>
            </div>

            <p v-if="review.comment" class="review-comment">{{ review.comment }}</p>

            <div v-if="review.images && review.images.length > 0" class="review-images">
              <img 
                v-for="(img, index) in review.images" 
                :key="index" 
                :src="img" 
                :alt="`Review image ${index + 1}`"
                class="review-image"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '../services/api'
import { authStore } from '../stores/authStore'

export default {
  name: 'TourDetail',
  setup() {
    const route = useRoute()
    const tourId = route.params.id

    const tour = ref({
      tags: [],
      name: '',
      description: '',
      status: '',
      difficulty: '',
      price: 0
    })
    const keyPoints = ref([])
    const reviews = ref([])
    
    const loading = ref(true)
    const loadingKeyPoints = ref(false)
    const loadingReviews = ref(false)
    const error = ref('')
    
    const showReviewForm = ref(false)
    const reviewLoading = ref(false)
    const imagesText = ref('')
    
    const reviewForm = ref({
      rating: 0,
      comment: '',
      visitedAt: null
    })

    const today = new Date().toISOString().split('T')[0]

    const isAuthor = computed(() => {
      const userId = authStore.getUserId()
      return userId === tour.value.authorId
    })

    const canAddReview = computed(() => {
      return authStore.isAuthenticated.value && !isAuthor.value
    })

    const fetchTour = async () => {
      loading.value = true
      error.value = ''

      try {
        const data = await api.getTourById(tourId)
        tour.value = {
          ...data,
          tags: data.tags || []
        }
      } catch (err) {
        error.value = err.response?.data?.error || err.message || 'Failed to load tour'
      } finally {
        loading.value = false
      }
    }

    const fetchKeyPoints = async () => {
      loadingKeyPoints.value = true
      try {
        const data = await api.getKeyPoints(tourId)
        keyPoints.value = data || []
      } catch (err) {
        console.error('Failed to load key points:', err)
        keyPoints.value = []
      } finally {
        loadingKeyPoints.value = false
      }
    }

    const fetchReviews = async () => {
      loadingReviews.value = true
      try {
        const data = await api.getReviews(tourId)
        reviews.value = data || []
      } catch (err) {
        console.error('Failed to load reviews:', err)
        reviews.value = []
      } finally {
        loadingReviews.value = false
      }
    }

    const submitReview = async () => {
      if (!reviewForm.value.rating) {
        alert('Please select a rating')
        return
      }

      reviewLoading.value = true

      try {
        const images = imagesText.value
          .split('\n')
          .map(url => url.trim())
          .filter(url => url.length > 0)

        const reviewData = {
          rating: reviewForm.value.rating,
          comment: reviewForm.value.comment,
          visitedAt: reviewForm.value.visitedAt ? new Date(reviewForm.value.visitedAt).toISOString() : null,
          images
        }

        await api.createReview(tourId, reviewData)
        
        // Reset form
        reviewForm.value = { rating: 0, comment: '', visitedAt: null }
        imagesText.value = ''
        showReviewForm.value = false
        
        // Reload reviews
        await fetchReviews()
      } catch (err) {
        alert(err.response?.data?.error || err.message || 'Failed to submit review')
      } finally {
        reviewLoading.value = false
      }
    }

    const cancelReview = () => {
      reviewForm.value = { rating: 0, comment: '', visitedAt: null }
      imagesText.value = ''
      showReviewForm.value = false
    }

    const formatDate = (dateString) => {
      return new Date(dateString).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      })
    }

    onMounted(async () => {
      await fetchTour()
      fetchKeyPoints()
      fetchReviews()
    })

    return {
      tour,
      keyPoints,
      reviews,
      loading,
      loadingKeyPoints,
      loadingReviews,
      error,
      isAuthor,
      canAddReview,
      showReviewForm,
      reviewForm,
      reviewLoading,
      imagesText,
      today,
      submitReview,
      cancelReview,
      formatDate
    }
  }
}
</script>

<style scoped>
.tour-detail {
  max-width: 1200px;
  margin: 2rem auto;
  padding: 2rem;
}

.loading,
.loading-small {
  text-align: center;
  padding: 2rem;
  color: #666;
}

.loading-small {
  padding: 1rem;
}

.error-message {
  background: #fee;
  color: #c33;
  padding: 1rem;
  border-radius: 8px;
  border: 1px solid #fcc;
}

.tour-header {
  margin-bottom: 2rem;
}

.tour-header h1 {
  color: #2c3e50;
  margin-bottom: 1rem;
}

.tour-meta {
  display: flex;
  gap: 1.5rem;
  align-items: center;
  flex-wrap: wrap;
}

.status-badge {
  padding: 0.4rem 1rem;
  border-radius: 20px;
  font-size: 0.9rem;
  font-weight: 600;
  text-transform: uppercase;
}

.status-badge.draft {
  background: #fff3e0;
  color: #f57c00;
}

.status-badge.published {
  background: #e8f5e9;
  color: #4caf50;
}

.difficulty,
.price {
  color: #666;
  font-weight: 500;
}

.tour-description {
  background: white;
  padding: 2rem;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  margin-bottom: 2rem;
}

.tour-description h2 {
  color: #2c3e50;
  margin-bottom: 1rem;
}

.tour-description p {
  color: #666;
  line-height: 1.6;
}

.tour-tags {
  margin-bottom: 2rem;
}

.tour-tags h3 {
  color: #2c3e50;
  margin-bottom: 0.5rem;
}

.tags-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.tag {
  background: #42b983;
  color: white;
  padding: 0.4rem 1rem;
  border-radius: 20px;
  font-size: 0.9rem;
}

.section {
  background: white;
  padding: 2rem;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  margin-bottom: 2rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.section-header h2 {
  color: #2c3e50;
  margin: 0;
}

.btn-add {
  background: #42b983;
  color: white;
  padding: 0.6rem 1.2rem;
  border-radius: 8px;
  border: none;
  text-decoration: none;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-add:hover {
  background: #35a372;
}

.empty-message {
  text-align: center;
  color: #999;
  padding: 2rem;
  font-style: italic;
}

.keypoints-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1.5rem;
}

.keypoint-card {
  border: 1px solid #ddd;
  border-radius: 8px;
  overflow: hidden;
  transition: all 0.2s;
}

.keypoint-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  border-color: #42b983;
}

.keypoint-image {
  width: 100%;
  height: 200px;
  object-fit: cover;
}

.keypoint-content {
  padding: 1rem;
}

.keypoint-content h3 {
  color: #2c3e50;
  margin: 0 0 0.5rem 0;
}

.keypoint-content p {
  color: #666;
  margin: 0 0 0.5rem 0;
}

.keypoint-location {
  color: #42b983;
  font-size: 0.9rem;
  font-weight: 500;
}

.review-form {
  background: #f8f9fa;
  padding: 1.5rem;
  border-radius: 8px;
  margin-bottom: 2rem;
}

.review-form h3 {
  color: #2c3e50;
  margin-bottom: 1rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 600;
  color: #333;
}

.rating-input {
  display: flex;
  gap: 0.5rem;
}

.star-btn {
  background: none;
  border: none;
  font-size: 2rem;
  color: #ddd;
  cursor: pointer;
  transition: color 0.2s;
}

.star-btn.active {
  color: #ffc107;
}

.star-btn:hover {
  color: #ffb300;
}

.form-group input[type="date"],
.form-group textarea {
  width: 100%;
  padding: 0.75rem;
  border: 2px solid #ddd;
  border-radius: 8px;
  font-size: 1rem;
}

.form-group textarea {
  resize: vertical;
}

.form-actions {
  display: flex;
  gap: 1rem;
  margin-top: 1rem;
}

.btn-primary,
.btn-secondary {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 8px;
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

.reviews-list {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.review-card {
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 1.5rem;
}

.review-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1rem;
}

.review-header h4 {
  color: #2c3e50;
  margin: 0 0 0.5rem 0;
}

.review-meta {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.rating {
  color: #ffc107;
  font-size: 1.1rem;
}

.visited-date,
.review-date {
  color: #999;
  font-size: 0.9rem;
}

.review-comment {
  color: #666;
  line-height: 1.6;
  margin: 0 0 1rem 0;
}

.review-images {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.review-image {
  width: 150px;
  height: 150px;
  object-fit: cover;
  border-radius: 8px;
}

@media (max-width: 768px) {
  .keypoints-grid {
    grid-template-columns: 1fr;
  }
}
</style>
