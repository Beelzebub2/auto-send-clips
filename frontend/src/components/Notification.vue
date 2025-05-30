<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

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

// Listen for video detection events
onMounted(() => {
  debugMessage.value = 'Component mounted, registering event listeners'
  console.log('Notification component mounted, setting up event listener for newVideoDetected')
  
  EventsOn('newVideoDetected', (data) => {
    debugMessage.value = `Received newVideoDetected: ${data?.fileName || 'unknown'}`
    console.log('Notification: New video detected event received:', data)
    if (data) {
      videoData.value = data
      customName.value = data.fileName || ''
      showNotification.value = true
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
  EventsOff('app-restored-from-tray')
})

function closeNotification() {
  showNotification.value = false
  customName.value = ''
  audioOnly.value = false
}

async function sendToDiscord() {
  if (!videoData.value.filePath) return
  
  sending.value = true
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
    alert('Failed to send to Discord: ' + error)
  } finally {
    sending.value = false
  }
}
</script>

<template>
  <div>
    <!-- Debug info always visible -->
    <div class="debug-bar">{{ debugMessage }}</div>
    
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
              placeholder="Enter custom message..."
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
  </div>
</template>

<style scoped>
.debug-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: rgba(0, 0, 0, 0.7);
  color: #ff8c00;
  padding: 4px 8px;
  font-size: 12px;
  z-index: 9999;
  text-align: center;
}

.notification-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.notification-modal {
  background: #2d2d2d;
  border-radius: 12px;
  border: 1px solid rgba(255, 140, 0, 0.3);
  width: 90%;
  max-width: 400px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
}

.notification-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid rgba(255, 140, 0, 0.2);
}

.notification-header h2 {
  color: #ff8c00;
  margin: 0;
  font-size: 1.1rem;
}

.close-btn {
  background: none;
  border: none;
  color: #ffffff;
  font-size: 1.5rem;
  cursor: pointer;
  padding: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn:hover {
  color: #ff8c00;
}

.notification-body {
  padding: 1.25rem;
}

.video-info {
  margin-bottom: 1rem;
  padding: 0.75rem;
  background: rgba(0, 0, 0, 0.3);
  border-radius: 6px;
  border-left: 3px solid #ff8c00;
}

.video-info p {
  margin: 0;
  color: #ffffff;
  font-size: 0.9rem;
  word-break: break-all;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: #ffffff;
  font-size: 0.9rem;
}

.form-input {
  width: 100%;
  padding: 0.6rem;
  border: 1px solid rgba(255, 140, 0, 0.3);
  border-radius: 6px;
  background: rgba(0, 0, 0, 0.3);
  color: #ffffff;
  font-size: 0.9rem;
}

.form-input:focus {
  outline: none;
  border-color: #ff8c00;
  background: rgba(0, 0, 0, 0.5);
}

.checkbox-label {
  display: flex;
  align-items: center;
  cursor: pointer;
}

.checkbox {
  margin-right: 0.5rem;
  accent-color: #ff8c00;
}

.button-group {
  display: flex;
  accent-color: #ff8c00;
}

.button-group {
  display: flex;
  gap: 0.75rem;
  margin-top: 1.5rem;
}

.btn-primary, .btn-secondary {
  flex: 1;
  padding: 0.75rem 1rem;
  border: none;
  border-radius: 6px;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-primary {
  background: #ff8c00;
  color: #1a1a1a;
}

.btn-primary:hover:not(:disabled) {
  background: #ff9933;
}

.btn-primary:disabled {
  background: #666;
  color: #999;
  cursor: not-allowed;
}

.btn-secondary {
  background: transparent;
  color: #ffffff;
  border: 1px solid rgba(255, 140, 0, 0.3);
}

.btn-secondary:hover {
  background: rgba(255, 140, 0, 0.1);
  border-color: rgba(255, 140, 0, 0.6);
}
</style>