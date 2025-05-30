<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { Clock, Eye, EyeOff, Folder, Activity } from 'lucide-vue-next'
import { GetAppStatus, StartMonitoring, StopMonitoring } from '../../wailsjs/go/main/App'

const status = ref({
  uptime: '0s',
  isMonitoring: false,
  monitorPath: '',
  lastActivity: 'Initializing...'
})

const isLoading = ref(true)
const error = ref('')
let statusInterval = null

const loadStatus = async () => {
  try {
    isLoading.value = true
    error.value = ''
    const appStatus = await GetAppStatus()
    status.value = appStatus
  } catch (err) {
    error.value = 'Failed to load status: ' + err.message
  } finally {
    isLoading.value = false
  }
}

const toggleMonitoring = async () => {
  try {
    if (status.value.isMonitoring) {
      await StopMonitoring()
    } else {
      await StartMonitoring()
    }
    await loadStatus()
  } catch (err) {
    error.value = 'Failed to toggle monitoring: ' + err.message
  }
}

onMounted(() => {
  loadStatus()
  // Update status every 2 seconds
  statusInterval = setInterval(loadStatus, 2000)
})

onUnmounted(() => {
  if (statusInterval) {
    clearInterval(statusInterval)
  }
})
</script>

<template>  <div class="status-page">
    <div class="status-header">
      <h2>Application Status</h2>
      <div class="status-indicator" :class="{ active: status.isMonitoring && !isLoading }">
        <div class="indicator-dot"></div>
        <span>{{ status.isMonitoring && !isLoading ? 'Monitoring Active' : 'Monitoring Inactive' }}</span>
      </div>
    </div>

    <div class="error-message" v-if="error">
      {{ error }}
    </div>

    <div class="status-grid" v-if="!isLoading">
      <div class="status-card">
        <div class="card-header">
          <h3>Uptime</h3>
          <Clock :size="20" class="icon" />
        </div>
        <div class="card-content">
          <div class="uptime-display">{{ status.uptime }}</div>
        </div>
      </div>

      <div class="status-card">
        <div class="card-header">
          <h3>Monitoring</h3>
          <component :is="status.isMonitoring ? Eye : EyeOff" :size="20" class="icon" :class="{ active: status.isMonitoring }" />
        </div>
        <div class="card-content">
          <div class="monitor-status">
            {{ status.isMonitoring ? 'Active' : 'Inactive' }}
          </div>
          <button 
            @click="toggleMonitoring" 
            class="toggle-button"
            :class="{ stop: status.isMonitoring }"
          >
            {{ status.isMonitoring ? 'Stop' : 'Start' }}
          </button>
        </div>
      </div>

      <div class="status-card full-width">
        <div class="card-header">
          <h3>Monitor Path</h3>
          <Folder :size="20" class="icon" />
        </div>
        <div class="card-content">
          <div class="monitor-path">{{ status.monitorPath || 'Not set' }}</div>
        </div>
      </div>

      <div class="status-card full-width">
        <div class="card-header">
          <h3>Last Activity</h3>
          <Activity :size="20" class="icon" />
        </div>
        <div class="card-content">
          <div class="last-activity">{{ status.lastActivity }}</div>
        </div>
      </div>
    </div>

    <div class="loading-spinner" v-if="isLoading">
      <div class="spinner"></div>
      <span>Loading status...</span>
    </div>
  </div>
</template>

<style scoped>
.status-page {
  padding: 1rem;
  height: calc(100vh - 60px);
  overflow: hidden;
}

.status-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid rgba(255, 140, 0, 0.3);
  flex-shrink: 0;
}

.status-header h2 {
  color: #ff8c00;
  font-size: 1.2rem;
  font-weight: 600;
  margin: 0;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 0.8rem;
  background: rgba(0, 0, 0, 0.3);
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  font-size: 0.8rem;
}

.status-indicator.active {
  background: rgba(0, 255, 0, 0.1);
  border-color: rgba(0, 255, 0, 0.3);
}

.indicator-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #666;
  animation: pulse 2s infinite;
}

.status-indicator.active .indicator-dot {
  background: #00ff00;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.error-message {
  background: rgba(255, 0, 0, 0.1);
  border: 1px solid rgba(255, 0, 0, 0.3);
  color: #ff6b6b;
  padding: 0.75rem;
  border-radius: 8px;
  margin-bottom: 1rem;
  font-size: 0.9rem;
  flex-shrink: 0;
}

.status-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.6rem;
  height: calc(100% - 80px);
}

.status-card {
  background: rgba(0, 0, 0, 0.3);
  border-radius: 6px;
  padding: 0.8rem;
  border: 1px solid rgba(255, 140, 0, 0.2);
  backdrop-filter: blur(10px);
  transition: all 0.3s ease;
  display: flex;
  flex-direction: column;
}

.status-card:hover {
  border-color: rgba(255, 140, 0, 0.4);
  transform: translateY(-1px);
}

.status-card.full-width {
  grid-column: 1 / -1;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
  flex-shrink: 0;
}

.card-header h3 {
  color: #ffffff;
  font-size: 0.95rem;
  font-weight: 500;
  margin: 0;
}

.icon {
  color: #ff8c00;
  flex-shrink: 0;
}

.icon.active {
  animation: bounce 1s infinite;
}

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-2px); }
}

.card-content {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  flex: 1;
  min-height: 0;
}

.uptime-display {
  font-size: 1.5rem;
  font-weight: 600;
  color: #ff8c00;
  text-align: center;
}

.monitor-status {
  font-size: 1rem;
  font-weight: 500;
  color: #ffffff;
  text-align: center;
}

.toggle-button {
  padding: 0.5rem 1rem;
  background: #ff8c00;
  border: none;
  border-radius: 6px;
  color: #1a1a1a;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-size: 0.8rem;
}

.toggle-button:hover {
  background: #e67e22;
  transform: translateY(-1px);
}

.toggle-button.stop {
  background: #e74c3c;
  color: #ffffff;
}

.toggle-button.stop:hover {
  background: #c0392b;
}

.monitor-path, .last-activity {
  font-size: 0.85rem;
  color: #cccccc;
  word-break: break-all;
  background: rgba(0, 0, 0, 0.2);
  padding: 0.6rem;
  border-radius: 6px;
  border-left: 3px solid #ff8c00;
  overflow: hidden;
  text-overflow: ellipsis;
}

.loading-spinner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  padding: 2rem;
  flex: 1;
  justify-content: center;
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid rgba(255, 140, 0, 0.3);
  border-top: 3px solid #ff8c00;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}
</style>
