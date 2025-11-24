<template>
  <div class="blog-create">
    <h1>Create New Blog</h1>
    
    <form @submit.prevent="handleSubmit">
      <div class="form-group">
        <label for="title">Title *</label>
        <input
          id="title"
          v-model="form.title"
          type="text"
          required
          placeholder="Enter blog title"
        />
      </div>

      <div class="form-group">
        <label for="description">Description (Markdown Supported) *</label>
        <div class="markdown-editor">
          <textarea
            id="description"
            v-model="form.description"
            required
            rows="15"
            placeholder="Write your blog content using markdown..."
          ></textarea>
          <div class="markdown-help">
            <p><strong>Markdown Tips:</strong></p>
            <ul>
              <li># Heading 1, ## Heading 2</li>
              <li>**bold**, *italic*</li>
              <li>[link](url)</li>
              <li>- list item</li>
            </ul>
          </div>
        </div>
      </div>

      <div class="form-group">
        <label for="images">Image URLs (optional, one per line)</label>
        <textarea
          id="images"
          v-model="imageUrls"
          rows="3"
          placeholder="https://example.com/image1.jpg&#10;https://example.com/image2.jpg"
        ></textarea>
      </div>

      <div v-if="error" class="error">{{ error }}</div>

      <div class="actions">
        <button type="button" @click="$router.back()" class="btn-cancel">Cancel</button>
        <button type="submit" :disabled="loading" class="btn-submit">
          {{ loading ? 'Creating...' : 'Create Blog' }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import api from '../services/api'

const router = useRouter()
const form = ref({
  title: '',
  description: '',
  images: []
})
const imageUrls = ref('')
const loading = ref(false)
const error = ref(null)

async function handleSubmit() {
  loading.value = true
  error.value = null
  
  try {
    // Parse image URLs
    const images = imageUrls.value
      .split('\n')
      .map(url => url.trim())
      .filter(url => url.length > 0)
    
    const blogData = {
      title: form.value.title,
      description: form.value.description,
      images: images
    }
    
    const createdBlog = await api.createBlog(blogData)
    router.push(`/blogs/${createdBlog.id}`)
  } catch (err) {
    console.error('Blog creation error:', err)
    error.value = err.response?.data?.error || err.response?.data || err.message || 'Failed to create blog'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.blog-create {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

h1 {
  margin-bottom: 30px;
  color: #2c3e50;
}

.form-group {
  margin-bottom: 25px;
}

label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #2c3e50;
}

input[type="text"],
textarea {
  width: 100%;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 5px;
  font-size: 14px;
  font-family: inherit;
}

input[type="text"]:focus,
textarea:focus {
  outline: none;
  border-color: #42b983;
}

.markdown-editor {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 15px;
}

.markdown-help {
  background: #f5f5f5;
  padding: 15px;
  border-radius: 5px;
  font-size: 13px;
}

.markdown-help p {
  margin: 0 0 10px 0;
  color: #2c3e50;
}

.markdown-help ul {
  margin: 0;
  padding-left: 20px;
}

.markdown-help li {
  color: #666;
  margin-bottom: 5px;
}

.error {
  color: #d32f2f;
  padding: 10px;
  background: #ffebee;
  border-radius: 5px;
  margin-bottom: 20px;
}

.actions {
  display: flex;
  gap: 15px;
  justify-content: flex-end;
}

.btn-cancel,
.btn-submit {
  padding: 10px 25px;
  border: none;
  border-radius: 5px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-cancel {
  background: #f5f5f5;
  color: #666;
}

.btn-cancel:hover {
  background: #e0e0e0;
}

.btn-submit {
  background: #42b983;
  color: white;
}

.btn-submit:hover:not(:disabled) {
  background: #35a372;
}

.btn-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
