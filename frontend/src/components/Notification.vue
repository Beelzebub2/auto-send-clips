<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { EventsOn, EventsOff, EventsEmit } from '../../wailsjs/runtime/runtime'
import ProgressModal from './ProgressModal.vue'

// Reactive data for notification
const showNotification = ref(false)
const videoData = ref({
  fileName: '',
  filePath: ''
})
const customName = ref('')
const audioOnly = ref(false)
const sending = ref(false)
const debugMessage = ref('Initializing...')

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

// Listen for video detection events
onMounted(() => {
  debugMessage.value = 'Component mounted, registering event listeners'
  console.log('Notification component mounted, setting up event listener for newVideoDetected')
  
  EventsOn('newVideoDetected', (data) => {
    debugMessage.value = `Received newVideoDetected: ${data?.fileName || 'unknown'}`
    console.log('Notification: New video detected event received:', data)
    if (data) {
      videoData.value = data
      customName.value = '' // Leave blank for user to fill
      showNotification.value = true
      console.log('Showing in-app notification modal')
    }
  })
  
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
  
  // Also listen for app restore event
  EventsOn('app-restored-from-tray', () => {
    console.log('Notification: App restored from tray')
    debugMessage.value = 'App restored from tray'
  })
  
  // Signal we're ready
  setTimeout(() => {
    debugMessage.value = 'Ready and listening for events'
  }, 1000)
})

onUnmounted(() => {
  console.log('Notification component unmounting, removing event listeners')
  EventsOff('newVideoDetected')
  EventsOff('sendProgress')
  EventsOff('compressionProgress')
  EventsOff('app-restored-from-tray')
})

function closeNotification() {
  showNotification.value = false
  customName.value = ''
  audioOnly.value = false
}

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

async function sendToDiscord() {
  if (!videoData.value.filePath) return
  
  sending.value = true
  
  // Delay showing progress modal to prevent jump scare
  setTimeout(() => {
    if (sending.value) { // Only show if still sending
      showProgress.value = true
    }
  }, 500)
  
  try {
    // Import the SendToDiscord method dynamically
    const { SendToDiscord } = await import('../../wailsjs/go/main/App')
    await SendToDiscord(videoData.value.filePath, customName.value, audioOnly.value)
    
    closeNotification()
    debugMessage.value = 'File sent to Discord successfully'
    console.log('File sent to Discord successfully')
  } catch (error) {
    console.error('Error sending to Discord:', error)
    debugMessage.value = `Error: ${error.message || error}`
    
    // Update progress with error
    progressData.value = {
      ...progressData.value,
      error: error.message || error.toString(),
      stage: 'error'
    }
  } finally {
    sending.value = false
  }
}
</script>

<template>  <div>
    <!-- Notification modal -->
    <div v-if="showNotification" class="notification-overlay">
      <div class="notification-modal">
        <div class="notification-header">
          <h2>🎬 New Video Detected!</h2>
          <button @click="closeNotification" class="close-btn">&times;</button>
        </div>
        
        <div class="notification-body">
          <div class="video-info">
            <p><strong>File:</strong> {{ videoData.fileName }}</p>
          </div>
            <div class="form-group">
            <label for="customName">Custom message (optional):</label>
            <input 
              id="customName" 
              v-model="customName" 
              type="text" 
              class="form-input"
              placeholder="Enter a custom name or message..."
            />
          </div>
          
          <div class="form-group">
            <label class="checkbox-label">
              <input 
                v-model="audioOnly" 
                type="checkbox" 
                class="checkbox"
              />
              Send audio only (extract audio from video)
            </label>
          </div>
          
          <div class="button-group">
            <button 
              @click="sendToDiscord" 
              :disabled="sending"
              class="btn-primary"
            >
              {{ sending ? 'Sending...' : 'Send to Discord' }}
            </button>
            <button @click="closeNotification" class="btn-secondary">
              Dismiss
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

<style scoped>
.notification-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(10, 13, 20, 0.9);
  backdrop-filter: blur(12px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  pointer-events: auto;
  animation: overlayFadeIn 0.3s ease-out;
}

@keyframes overlayFadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.notification-modal {
  background: var(--bg-cards);
  border-radius: 20px;
  border: 1px solid var(--border-default);
  width: 480px;
  max-width: 90vw;
  box-shadow: var(--shadow-xl);
  pointer-events: auto;
  animation: modalSlideIn 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  backdrop-filter: blur(20px);
  position: relative;
  overflow: hidden;
}

.notification-modal::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, var(--primary-color), var(--primary-light));
  opacity: 1;
}

.notification-modal::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, 
    rgba(255, 124, 61, 0.03) 0%, 
    transparent 50%, 
    rgba(255, 124, 61, 0.01) 100%);
  pointer-events: none;
}

@keyframes modalSlideIn {
  from {
    transform: scale(0.8) translateY(40px);
    opacity: 0;
  }
  to {
    transform: scale(1) translateY(0);
    opacity: 1;
  }
}

.notification-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border-default);
  background: linear-gradient(135deg, var(--bg-elements), var(--bg-cards));
  position: relative;
  z-index: 1;
}

.notification-header h2 {
  color: var(--primary-color);
  margin: 0;
  font-size: 1.2rem;
  font-weight: 700;
  text-shadow: 0 2px 12px rgba(255, 124, 61, 0.4);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.close-btn {
  background: var(--bg-interactive);
  border: 1px solid var(--border-default);
  color: var(--text-secondary);
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0.4rem;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  transition: var(--transition-smooth);
  font-weight: 600;
}

.close-btn:hover {
  color: var(--primary-color);
  background: var(--bg-elements);
  border-color: var(--border-accent);
  transform: scale(1.05);
  box-shadow: var(--shadow-small);
}

.notification-body {
  padding: 2rem 1.5rem;
  position: relative;
  z-index: 1;
}

.video-info {
  margin-bottom: 2rem;
  padding: 1.25rem;
  background: linear-gradient(135deg, var(--bg-elements), var(--bg-interactive));
  border-radius: 12px;
  border: 1px solid var(--border-default);
  border-left: 4px solid var(--primary-color);
  box-shadow: var(--shadow-small);
  position: relative;
  overflow: hidden;
}

.video-info::before {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  width: 60px;
  height: 60px;
  background: radial-gradient(circle, rgba(255, 124, 61, 0.1) 0%, transparent 70%);
  border-radius: 50%;
  transform: translate(20px, -20px);
}

.video-info p {
  margin: 0;
  color: var(--text-primary);
  font-size: 0.95rem;
  word-break: break-all;
  font-weight: 500;
  position: relative;
  z-index: 1;
}

.video-info strong {
  color: var(--primary-color);
  text-shadow: 0 1px 4px rgba(255, 124, 61, 0.3);
}

.form-group {
  margin-bottom: 2rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.75rem;
  color: var(--text-primary);
  font-size: 0.95rem;
  font-weight: 600;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
}

.form-input {
  width: 100%;
  padding: 1rem;
  border: 1px solid var(--border-default);
  border-radius: 12px;
  background: var(--bg-darkest);
  color: var(--text-primary);
  font-size: 0.95rem;
  transition: var(--transition-smooth);
  box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.2), var(--shadow-small);
  font-weight: 500;
}

.form-input:focus {
  outline: none;
  border-color: var(--primary-color);
  background: linear-gradient(135deg, var(--bg-elements), var(--bg-darkest));
  box-shadow: 0 0 0 3px rgba(255, 124, 61, 0.15), 
              inset 0 1px 3px rgba(0, 0, 0, 0.1),
              var(--shadow-medium);
  transform: translateY(-1px);
}

.form-input::placeholder {
  color: var(--text-muted);
  font-style: italic;
}

.checkbox-label {
  display: flex;
  align-items: center;
  cursor: pointer;
  padding: 0.75rem;
  border-radius: 12px;
  transition: var(--transition-smooth);
  border: 1px solid transparent;
  background: linear-gradient(135deg, var(--bg-elements), var(--bg-cards));
  font-weight: 500;
}

.checkbox-label:hover {
  background: linear-gradient(135deg, var(--bg-interactive), var(--bg-elements));
  border-color: var(--border-light);
  transform: translateY(-1px);
  box-shadow: var(--shadow-small);
}

.checkbox {
  margin-right: 1rem;
  accent-color: var(--primary-color);
  transform: scale(1.3);
  cursor: pointer;
}

.button-group {
  display: flex;
  gap: 1rem;
  margin-top: 2.5rem;
}

.btn-primary, .btn-secondary {
  flex: 1;
  padding: 1rem 1.5rem;
  border: none;
  border-radius: 12px;
  font-size: 0.95rem;
  font-weight: 700;
  cursor: pointer;
  transition: var(--transition-smooth);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  box-shadow: var(--shadow-small);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  position: relative;
  overflow: hidden;
}

.btn-primary::before,
.btn-secondary::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.1), transparent);
  transition: left 0.6s ease;
}

.btn-primary:hover::before,
.btn-secondary:hover::before {
  left: 100%;
}

.btn-primary {
  background: linear-gradient(135deg, var(--primary-color), var(--primary-dark));
  color: #ffffff;
  border: 1px solid var(--primary-color);
  box-shadow: 0 4px 16px rgba(255, 124, 61, 0.3);
}

.btn-primary:hover:not(:disabled) {
  background: linear-gradient(135deg, var(--primary-light), var(--primary-color));
  transform: translateY(-3px);
  box-shadow: 0 8px 24px rgba(255, 124, 61, 0.4);
}

.btn-primary:disabled {
  background: linear-gradient(135deg, var(--bg-elements), var(--bg-interactive));
  color: var(--text-muted);
  cursor: not-allowed;
  transform: none;
  box-shadow: var(--shadow-small);
  opacity: 0.6;
}

.btn-secondary {
  background: linear-gradient(135deg, var(--bg-elements), var(--bg-interactive));
  color: var(--text-primary);
  border: 1px solid var(--border-default);
}

.btn-secondary:hover {
  background: linear-gradient(135deg, var(--bg-interactive), var(--bg-elements));
  border-color: var(--border-light);
  transform: translateY(-3px);
  box-shadow: var(--shadow-medium);
  color: var(--primary-color);
}

/* Mobile responsive */
@media (max-width: 768px) {
  .notification-modal {
    width: 95%;
    max-width: 420px;
    margin: 1rem;
    border-radius: 16px;
  }
  
  .notification-header {
    padding: 1rem 1.25rem;
  }
  
  .notification-header h2 {
    font-size: 1.1rem;
  }
  
  .notification-body {
    padding: 1.5rem 1.25rem;
  }
  
  .button-group {
    flex-direction: column;
    gap: 0.75rem;
    margin-top: 2rem;
  }
  
  .btn-primary, .btn-secondary {
    width: 100%;
    padding: 0.875rem 1.25rem;
  }
}

@media (max-width: 480px) {
  .notification-modal {
    width: 98%;
    margin: 0.5rem;
  }
  
  .notification-header {
    padding: 0.875rem 1rem;
  }
  
  .notification-body {
    padding: 1.25rem 1rem;
  }
  
  .video-info {
    padding: 1rem;
  }
  
  .form-input {
    padding: 0.875rem;
  }
  
  .checkbox-label {
    padding: 0.625rem;
  }
}
</style>