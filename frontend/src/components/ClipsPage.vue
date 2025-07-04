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
            <div v-for="clip in clips" :key="clip.uuid" class="clip-card">
                <div class="clip-thumbnail">
                    <div class="thumbnail-container">
                        <div class="thumbnail-placeholder" :class="{ 'loaded': imageStates[clip.uuid]?.loaded }">
                            <div class="placeholder-icon">
                                <svg viewBox="0 0 24 24" fill="currentColor">
                                    <path
                                        d="M21 19V5c0-1.1-.9-2-2-2H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2zM8.5 13.5l2.5 3.01L14.5 12l4.5 6H5l3.5-4.5z" />
                                </svg>
                            </div>
                            <div v-if="imageStates[clip.uuid]?.loading" class="thumbnail-spinner"></div>
                        </div>
                        <img :src="getThumbnailSrc(clip)" :alt="clip.title"
                            :class="{ 'loaded': imageStates[clip.uuid]?.loaded, 'error': imageStates[clip.uuid]?.error }"
                            @load="onImageLoad(clip.uuid)" @error="onImageError(clip.uuid)" />
                    </div>
                    <div class="clip-duration" v-if="clip.duration">
                        {{ formatDuration(clip.duration) }}
                    </div>
                </div>

                <div class="clip-info">
                    <h3 class="clip-title">{{ clip.title }}</h3>
                    <p class="clip-game" v-if="clip.gameTitle">{{ clip.gameTitle }}</p>
                    <p class="clip-date">{{ formatDate(clip.timeCreated) }}</p>

                    <div class="clip-actions">
                        <button @click="sendToDiscord(clip.uuid)" :disabled="sendingClips.has(clip.uuid)"
                            class="send-button">
                            <span v-if="sendingClips.has(clip.uuid)">Sending...</span>
                            <span v-else>Send to Discord</span>
                        </button>
                    </div>
                </div>
            </div>
        </div>
        
    <!-- Progress Modal -->
    <ProgressModal 
      :isVisible="showProgress"
      :title="progressData.title"
      :progress="progressData.progress"
      :stage="progressData.stage"
      :message="progressData.message"
      :error="progressData.error"
      :isComplete="progressData.isComplete"
      @close="closeProgress"
    />
    </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { GetMedalTVClips, SendClipToDiscord } from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import ProgressModal from './ProgressModal.vue'

const clips = ref([])
const isLoading = ref(true)
const error = ref('')
const sendingClips = ref(new Set())
const imageStates = ref({})

// Progress modal data
const showProgress = ref(false)
const progressData = ref({
  title: 'Sending to Discord',
  progress: 0,
  stage: '',
  message: '',
  error: '',
  isComplete: false
})

// Listen for progress events
onMounted(() => {
  // Listen for send progress events
  EventsOn('sendProgress', (data) => {
    console.log('Send progress:', data)
    progressData.value = {
      title: 'Sending to Discord',
      progress: data.progress || 0,
      stage: data.stage || '',
      message: data.message || '',
      error: data.error || '',
      isComplete: data.isComplete || false
    }
  })
  
  // Listen for compression progress events
  EventsOn('compressionProgress', (data) => {
    console.log('Compression progress:', data)
    progressData.value = {
      title: 'Compressing Video',
      progress: data.Progress || 0,
      stage: data.Stage || '',
      message: data.Message || '',
      error: data.Error || '',
      isComplete: data.IsComplete || false
    }
  })
  
  loadClips()
})

onUnmounted(() => {
  EventsOff('sendProgress')
  EventsOff('compressionProgress')
})

function closeProgress() {
  showProgress.value = false
  progressData.value = {
    title: 'Processing',
    progress: 0,
    stage: '',
    message: '',
    error: '',
    isComplete: false
  }
}

const loadClips = async () => {
    isLoading.value = true
    error.value = ''

    try {
        const clipsData = await GetMedalTVClips()
        clips.value = clipsData || []

        // Initialize image states for each clip
        imageStates.value = {}
        clips.value.forEach(clip => {
            imageStates.value[clip.uuid] = {
                loading: true,
                loaded: false,
                error: false
            }
        })
    } catch (err) {
        error.value = err.message || 'Failed to load clips'
        console.error('Error loading clips:', err)
    } finally {
        isLoading.value = false
    }
}

const sendToDiscord = async (clipUUID) => {
    sendingClips.value.add(clipUUID)
    
    // Delay showing progress modal to prevent jump scare
    setTimeout(() => {
        if (sendingClips.value.has(clipUUID)) { // Only show if still sending
            showProgress.value = true
        }
    }, 500)

    try {
        await SendClipToDiscord(clipUUID)
        // Show success notification or update UI
        console.log('Clip sent successfully')
    } catch (err) {
        console.error('Error sending clip:', err)
        
        // Update progress with error
        progressData.value = {
            ...progressData.value,
            error: err.message || err.toString(),
            stage: 'error'
        }
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

const onImageLoad = (clipUUID) => {
    if (imageStates.value[clipUUID]) {
        imageStates.value[clipUUID].loading = false
        imageStates.value[clipUUID].loaded = true
        imageStates.value[clipUUID].error = false
    }
}

const onImageError = (clipUUID) => {
    if (imageStates.value[clipUUID]) {
        imageStates.value[clipUUID].loading = false
        imageStates.value[clipUUID].loaded = false
        imageStates.value[clipUUID].error = true
    }
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
    display: flex;
    justify-content: center;
    align-items: center;
    margin-bottom: 1.5rem;
    padding-bottom: 1rem;
    border-bottom: 1px solid var(--border-default);
    flex-shrink: 0;
}

.clips-header h2 {
    color: var(--primary-color);
    font-size: 1.3rem;
    font-weight: 700;
    margin: 0;
    text-shadow: 0 2px 8px rgba(255, 124, 61, 0.3);
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
    0% {
        transform: rotate(0deg);
    }

    100% {
        transform: rotate(360deg);
    }
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

.thumbnail-container {
    position: relative;
    width: 100%;
    height: 100%;
}

.thumbnail-placeholder {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: linear-gradient(135deg, #2a2a2a 0%, #1a1a1a 100%);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    transition: opacity 0.3s ease;
    z-index: 1;
    animation: placeholder-pulse 2s ease-in-out infinite;
}

@keyframes placeholder-pulse {

    0%,
    100% {
        background: linear-gradient(135deg, #2a2a2a 0%, #1a1a1a 100%);
    }

    50% {
        background: linear-gradient(135deg, #2f2f2f 0%, #1f1f1f 100%);
    }
}

.thumbnail-placeholder.loaded {
    opacity: 0;
    pointer-events: none;
}

.placeholder-icon {
    width: 48px;
    height: 48px;
    color: #555;
    margin-bottom: 8px;
}

.thumbnail-spinner {
    width: 24px;
    height: 24px;
    border: 2px solid #333;
    border-top: 2px solid var(--primary-color);
    border-radius: 50%;
    animation: spin 1s linear infinite;
}

.clip-thumbnail img {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
    transition: opacity 0.3s ease, transform 0.3s ease;
    opacity: 0;
    z-index: 2;
}

.clip-thumbnail img.loaded {
    opacity: 1;
}

.clip-thumbnail img.error {
    display: none;
}

.clip-card:hover .clip-thumbnail img.loaded {
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
