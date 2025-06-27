<template>
  <div class="clips-page">
    <div class="clips-header">
      <h2>Medal TV Clips</h2>
    </div>

    <div v-if="isLoading" class="loading-state">
      <div class="loading-spinner"></div>
      <p>Loading clips...</p>
    </div>

    <div v-else-if="error" class="error-state">
      <p class="error-message">{{ error }}</p>
      <button @click="loadClips" class="retry-button">
        Retry
      </button>
    </div>

    <div v-else-if="clips.length === 0" class="empty-state">
      <p>No clips found</p>
      <p class="empty-subtitle">Make sure Medal TV is installed and has recorded clips.</p>
    </div>

    <div v-else class="clips-grid">
      <div 
        v-for="clip in clips" 
        :key="clip.uuid"
        class="clip-card"
      >
        <div class="clip-thumbnail">
          <img 
            :src="getThumbnailSrc(clip)" 
            :alt="clip.title"
            @error="onImageError"
          />
          <div class="clip-duration" v-if="clip.duration">
            {{ formatDuration(clip.duration) }}
          </div>
        </div>
        
        <div class="clip-info">
          <h3 class="clip-title">{{ clip.title }}</h3>
          <p class="clip-game" v-if="clip.gameTitle">{{ clip.gameTitle }}</p>
          <p class="clip-date">{{ formatDate(clip.timeCreated) }}</p>
          
          <div class="clip-actions">
            <button 
              @click="sendToDiscord(clip.uuid)"
              :disabled="sendingClips.has(clip.uuid)"
              class="send-button"
            >
              <span v-if="sendingClips.has(clip.uuid)">Sending...</span>
              <span v-else>Send to Discord</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { GetMedalTVClips, SendClipToDiscord } from '../../wailsjs/go/main/App'

const clips = ref([])
const isLoading = ref(true)
const error = ref('')
const sendingClips = ref(new Set())

const loadClips = async () => {
  isLoading.value = true
  error.value = ''
  
  try {
    const clipsData = await GetMedalTVClips()
    clips.value = clipsData || []
  } catch (err) {
    error.value = err.message || 'Failed to load clips'
    console.error('Error loading clips:', err)
  } finally {
    isLoading.value = false
  }
}

const sendToDiscord = async (clipUUID) => {
  sendingClips.value.add(clipUUID)
  
  try {
    await SendClipToDiscord(clipUUID)
    // Show success notification or update UI
    console.log('Clip sent successfully')
  } catch (err) {
    console.error('Error sending clip:', err)
    alert(`Failed to send clip: ${err.message}`)
  } finally {
    sendingClips.value.delete(clipUUID)
  }
}

const formatDuration = (seconds) => {
  if (!seconds) return '0:00'
  
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = Math.floor(seconds % 60)
  return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`
}

const formatDate = (timestamp) => {
  if (!timestamp) return ''
  
  const date = new Date(timestamp * 1000)
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { 
    hour: '2-digit', 
    minute: '2-digit' 
  })
}

const getThumbnailSrc = (clip) => {
  // First try the online thumbnail URL if available
  if (clip.thumbnailUrl) {
    return clip.thumbnailUrl
  }
  
  // Fallback to local thumbnail with proper file:// protocol
  if (clip.thumbnail) {
    return `file:///${clip.thumbnail.replace(/\\/g, '/')}`
  }
  
  // If no thumbnail available, return empty string (will trigger onImageError)
  return ''
}

const onImageError = (event) => {
  // Hide broken images or show placeholder
  event.target.style.display = 'none'
}

onMounted(() => {
  loadClips()
})
</script>

<style scoped>
.clips-page {
  padding: 1.5rem;
  height: 100%;
  overflow-y: auto;
}

.clips-header {
  margin-bottom: 2rem;
}

.clips-header h2 {
  color: var(--text-primary);
  margin-bottom: 0;
  font-size: 1.8rem;
  font-weight: 600;
}

.loading-state,
.error-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 300px;
  text-align: center;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border-default);
  border-top: 3px solid var(--primary-color);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.error-message {
  color: var(--error-color);
  margin-bottom: 1rem;
}

.retry-button {
  padding: 0.5rem 1rem;
  background: var(--primary-color);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: var(--transition-smooth);
}

.retry-button:hover {
  background: var(--primary-dark);
}

.empty-subtitle {
  color: var(--text-secondary);
  font-size: 0.9rem;
  margin-top: 0.5rem;
}

.clips-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1.5rem;
  padding-bottom: 2rem;
}

.clip-card {
  background: var(--bg-sections);
  border: 1px solid var(--border-default);
  border-radius: 12px;
  overflow: hidden;
  transition: var(--transition-smooth);
  box-shadow: var(--shadow-small);
}

.clip-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-medium);
  border-color: var(--border-accent);
}

.clip-thumbnail {
  position: relative;
  width: 100%;
  height: 180px;
  overflow: hidden;
  background: var(--bg-darkest);
}

.clip-thumbnail img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: var(--transition-smooth);
}

.clip-card:hover .clip-thumbnail img {
  transform: scale(1.05);
}

.clip-duration {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.8);
  color: white;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
}

.clip-info {
  padding: 1rem;
}

.clip-title {
  color: var(--text-primary);
  font-size: 1rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
  line-height: 1.3;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.clip-game {
  color: var(--text-secondary);
  font-size: 0.85rem;
  margin-bottom: 0.25rem;
}

.clip-date {
  color: var(--text-tertiary);
  font-size: 0.75rem;
  margin-bottom: 1rem;
}

.clip-actions {
  display: flex;
  gap: 0.5rem;
}

.send-button {
  flex: 1;
  padding: 0.6rem 1rem;
  background: linear-gradient(135deg, var(--primary-color), var(--primary-dark));
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 0.85rem;
  font-weight: 500;
  transition: var(--transition-smooth);
  box-shadow: var(--shadow-small);
}

.send-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: var(--shadow-medium);
}

.send-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .clips-grid {
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 1rem;
  }
  
  .clips-page {
    padding: 1rem;
  }
}

@media (max-width: 480px) {
  .clips-grid {
    grid-template-columns: 1fr;
  }
  
  .clip-thumbnail {
    height: 160px;
  }
}
</style>
