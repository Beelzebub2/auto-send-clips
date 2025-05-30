<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { Clock, Folder, Activity } from 'lucide-vue-next'
import { GetAppStatus, StartMonitoring, StopMonitoring } from '../../wailsjs/go/main/App'

const status = ref({
  uptime: '0s',
  isMonitoring: false,
  monitorPath: '',
  videosSent: 0,
  audiosSent: 0
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
      </div>      <div class="status-card">
        <div class="card-header">
          <h3>Monitoring</h3>
          <div class="monitoring-indicator" :class="{ active: status.isMonitoring }">
            <div v-if="status.isMonitoring" class="ping-animation">
              <div class="ping-dot"></div>
              <div class="ping-wave"></div>
            </div>
            <div v-else class="inactive-dot"></div>
          </div>
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
      </div>      <div class="status-card full-width">
        <div class="card-header">
          <h3>Total Clips Sent</h3>
          <Activity :size="20" class="icon" />
        </div>
        <div class="card-content">
          <div class="clips-stats">
            <div class="stat-item">
              <span class="stat-label">Videos:</span>
              <span class="stat-value">{{ status.videosSent }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">Audios:</span>
              <span class="stat-value">{{ status.audiosSent }}</span>
            </div>
          </div>
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

.monitoring-indicator {
  position: relative;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.ping-animation {
  position: relative;
  width: 20px;
  height: 20px;
}

.ping-dot {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 8px;
  height: 8px;
  background: #00ff00;
  border-radius: 50%;
  transform: translate(-50%, -50%);
  z-index: 2;
}

.ping-wave {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 8px;
  height: 8px;
  background: #00ff00;
  border-radius: 50%;
  transform: translate(-50%, -50%);
  animation: ping 1s cubic-bezier(0, 0, 0.2, 1) infinite;
  opacity: 0.75;
}

@keyframes ping {
  75%, 100% {
    transform: translate(-50%, -50%) scale(2.5);
    opacity: 0;
  }
}

.inactive-dot {
  width: 8px;
  height: 8px;
  background: #666;
  border-radius: 50%;
  opacity: 0.6;
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
  color: #ccc;
  font-family: 'Courier New', monospace;
  font-size: 0.85rem;
  word-break: break-all;
}

.clips-stats {
  display: flex;
  gap: 1.5rem;
  justify-content: center;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.25rem;
}

.stat-label {
  color: #999;
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stat-value {
  color: #ff8c00;
  font-size: 1.5rem;
  font-weight: 700;
  font-family: 'Courier New', monospace;
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
  position: relative;
  width: 40px;
  height: 40px;
}

.spinner::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 12px;
  height: 12px;
  background: #ff8c00;
  border-radius: 50%;
  transform: translate(-50%, -50%);
  z-index: 2;
}

.spinner::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 12px;
  height: 12px;
  background: #ff8c00;
  border-radius: 50%;
  transform: translate(-50%, -50%);
  animation: spinnerPing 1.5s cubic-bezier(0, 0, 0.2, 1) infinite;
  opacity: 0.6;
}

@keyframes spinnerPing {
  75%, 100% {
    transform: translate(-50%, -50%) scale(3);
    opacity: 0;
  }
}
</style>
