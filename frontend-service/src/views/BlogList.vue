<template>
  <div class="blog-list">
    <div class="header">
      <h1>Blog Feed</h1>
      <button @click="$router.push('/blogs/create')" class="create-btn">Create Blog</button>
    </div>

    <div v-if="loading" class="loading">Loading blogs...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="blogs.length === 0" class="empty">
      No blogs yet. Follow some users or create your first blog!
    </div>
    <div v-else class="blogs">
      <div v-for="blog in blogs" :key="blog.id" class="blog-card" @click="goToBlog(blog.id)">
        <div class="blog-header">
          <h2>{{ blog.title }}</h2>
          <span class="date">{{ formatDate(blog.created_at) }}</span>
        </div>
        <p class="author">by {{ blog.author_name }}</p>
        <p class="description">{{ getPreview(blog.description) }}</p>
        <div class="blog-footer">
          <span class="likes">❤️ {{ blog.likes_count || 0 }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '../services/api'

const router = useRouter()
const blogs = ref([])
const loading = ref(true)
const error = ref(null)

onMounted(async () => {
  await fetchBlogs()
})

async function fetchBlogs() {
  loading.value = true
  error.value = null
  try {
    const result = await api.getBlogs()
    blogs.value = result || []
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to load blogs'
    blogs.value = []
  } finally {
    loading.value = false
  }
}

function goToBlog(id) {
  router.push(`/blogs/${id}`)
}

function formatDate(dateString) {
  return new Date(dateString).toLocaleDateString('en-US', { 
    year: 'numeric', 
    month: 'long', 
    day: 'numeric' 
  })
}

function getPreview(description) {
  // Strip markdown syntax for preview
  const plain = description.replace(/[#*_~`\[\]]/g, '')
  return plain.length > 200 ? plain.substring(0, 200) + '...' : plain
}
</script>

<style scoped>
.blog-list {
  max-width: 900px;
  margin: 0 auto;
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.create-btn {
  background-color: #42b983;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 5px;
  cursor: pointer;
  font-size: 14px;
}

.create-btn:hover {
  background-color: #35a372;
}

.loading, .error, .empty {
  text-align: center;
  padding: 40px;
  color: #666;
}

.error {
  color: #d32f2f;
}

.blogs {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.blog-card {
  background: white;
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 20px;
  cursor: pointer;
  transition: all 0.2s;
}

.blog-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  border-color: #42b983;
}

.blog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.blog-header h2 {
  margin: 0;
  color: #2c3e50;
  font-size: 24px;
}

.date {
  color: #666;
  font-size: 14px;
}

.author {
  color: #42b983;
  font-weight: 500;
  margin: 5px 0 15px 0;
}

.description {
  color: #333;
  line-height: 1.6;
  margin: 15px 0;
}

.blog-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px solid #eee;
}

.likes {
  font-size: 16px;
  color: #666;
}
</style>
