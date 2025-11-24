<template>
  <div class="home-container">
    <div class="hero-section">
      <h1 class="hero-title">Hello</h1>
      <p class="hero-subtitle">Welcome to our platform</p>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Home'
}
</script>

<style scoped>
.home-container {
  min-height: calc(100vh - 80px);
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.hero-section {
  text-align: center;
  color: white;
}

.hero-title {
  font-size: 5rem;
  font-weight: 300;
  margin: 0 0 1rem 0;
  animation: fadeIn 1s ease-in;
}

.hero-subtitle {
  font-size: 1.5rem;
  font-weight: 300;
  opacity: 0.9;
  animation: fadeIn 1.5s ease-in;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 768px) {
  .hero-title {
    font-size: 3rem;
  }

  .hero-subtitle {
    font-size: 1.2rem;
  }
}
</style>
