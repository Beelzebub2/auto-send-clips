<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { Clock, Folder, Activity } from 'lucide-vue-next'
import { GetAppStatus, StartMonitoring, StopMonitoring } from '../../wailsjs/go/main/App'

const status = ref({
  uptime: '0s',
  isMonitoring: false,
  monitorPath: '',
  videosSent: 0,
  audiosSent: 0,
  useMedalTV: false,
  useNVIDIA: false,
  useCustom: false,
  medalTVPath: '',
  nvidiaPath: ''
})

const isLoading = ref(true)
const error = ref('')
let statusInterval = null

const loadStatus = async (showLoading = true) => {
  try {
    if (showLoading) {
      isLoading.value = true
    }
    error.value = ''
    const appStatus = await GetAppStatus()
    status.value = appStatus
  } catch (err) {
    error.value = 'Failed to load status: ' + err.message
  } finally {
    if (showLoading) {
      isLoading.value = false
    }
  }
}

const updateStatusSilently = async () => {
  try {
    const appStatus = await GetAppStatus()
    status.value = appStatus
  } catch (err) {
    // Silently fail on background updates to avoid disrupting the UI
    console.warn('Background status update failed:', err.message)
  }
}

const toggleMonitoring = async () => {
  try {
    if (status.value.isMonitoring) {
      await StopMonitoring()
    } else {
      await StartMonitoring()
    }
    await loadStatus(false) // Don't show loading on toggle
  } catch (err) {
    error.value = 'Failed to toggle monitoring: ' + err.message
  }
}

onMounted(() => {
  loadStatus()
  // Update status every 2 seconds without showing loading
  statusInterval = setInterval(updateStatusSilently, 2000)
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
      </div>      <div class="status-card full-width">
        <div class="card-header">
          <h3>Monitor Paths</h3>
          <Folder :size="20" class="icon" />
        </div>        <div class="card-content">
          <div class="paths-container">
            <!-- Show MedalTV path if enabled -->
            <div class="monitor-path medaltu-path" v-if="status.useMedalTV && status.medalTVPath">
              <div class="path-label">
                <div class="path-icon medaltu-icon">M</div>
                MedalTV:
              </div>
              <div class="path-value">{{ status.medalTVPath }}</div>
            </div>
            
            <!-- Show NVIDIA path if enabled -->
            <div class="monitor-path nvidia-path" v-if="status.useNVIDIA && status.nvidiaPath">
              <div class="path-label">
                <div class="path-icon nvidia-icon">N</div>
                NVIDIA:
              </div>
              <div class="path-value">{{ status.nvidiaPath }}</div>
            </div>
            
            <!-- Show custom path if enabled -->
            <div class="monitor-path custom-path" v-if="status.useCustom && status.monitorPath">
              <div class="path-label">
                <div class="path-icon custom-icon">C</div>
                Custom:
              </div>
              <div class="path-value">{{ status.monitorPath }}</div>
            </div>
              <!-- Show fallback message if no paths -->
            <div class="monitor-path no-paths" v-if="!status.useMedalTV && !status.useNVIDIA && !status.useCustom">
              <div class="path-value">No monitoring paths configured</div>
            </div>
          </div>
        </div>
      </div><div class="status-card full-width">
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
  padding: 1.5rem;
  height: 100vh;
  overflow: auto;
  background: #0a0d14;
  display: flex;
  flex-direction: column;
}

.status-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #3a4553;
  flex-shrink: 0;
}

.status-header h2 {
  color: #ff7c3d;
  font-size: 1.3rem;
  font-weight: 700;
  margin: 0;
  text-shadow: 0 2px 8px rgba(255, 124, 61, 0.3);
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.5rem 1rem;
  background: #2a3441;
  border-radius: 20px;
  border: 1px solid #3a4553;
  font-size: 0.85rem;
  font-weight: 500;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.3);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  opacity: 0;
  transform: translateY(10px);
  animation: fadeInUp 0.6s cubic-bezier(0.4, 0, 0.2, 1) forwards;
}

.status-indicator.active {
  background: linear-gradient(135deg, rgba(74, 222, 128, 0.1), rgba(74, 222, 128, 0.05));
  border-color: rgba(74, 222, 128, 0.4);
  color: #4ade80;
  box-shadow: 0 4px 16px rgba(74, 222, 128, 0.2);
}

@keyframes fadeInUp {
  0% {
    opacity: 0;
    transform: translateY(10px);
  }
  100% {
    opacity: 1;
    transform: translateY(0);
  }
}

.indicator-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #94a3b8;
  animation: pulse-glow 2s infinite;
}

.status-indicator.active .indicator-dot {
  background: #4ade80;
}

@keyframes pulse-glow {
  0%, 100% { 
    opacity: 1;
    transform: scale(1);
  }
  50% { 
    opacity: 0.6;
    transform: scale(1.1);
  }
}

.error-message {
  background: linear-gradient(135deg, rgba(255, 107, 107, 0.1), rgba(255, 107, 107, 0.05));
  border: 1px solid rgba(255, 107, 107, 0.3);
  color: #ff6b6b;
  padding: 1rem;
  border-radius: 12px;
  margin-bottom: 1.5rem;
  font-size: 0.9rem;
  flex-shrink: 0;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.3);
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1rem;
  flex: 1;
  align-content: start;
  grid-auto-rows: min-content;
}

.status-card {
  background: #1e2530;
  border-radius: 12px;
  padding: 1.2rem;
  border: 1px solid #3a4553;
  backdrop-filter: blur(20px);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  flex-direction: column;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.3);
  position: relative;
  overflow: hidden;
  opacity: 0;
  transform: translateY(20px) scale(0.95);
  animation: cardEntrance 0.6s cubic-bezier(0.4, 0, 0.2, 1) forwards;
  height: fit-content;
}

.status-card:nth-child(1) { animation-delay: 0.1s; }
.status-card:nth-child(2) { animation-delay: 0.2s; }
.status-card:nth-child(3) { animation-delay: 0.3s; }
.status-card:nth-child(4) { animation-delay: 0.4s; }

@keyframes cardEntrance {
  0% {
    opacity: 0;
    transform: translateY(20px) scale(0.95);
  }
  100% {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.status-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, #ff7c3d, #ff9966);
  opacity: 0;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.status-card:hover {
  border-color: #4a5563;
  transform: translateY(-4px) scale(1.01);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4), 0 0 24px rgba(255, 124, 61, 0.1);
  background: #2a3441;
}

.status-card:hover::before {
  opacity: 1;
}

.status-card.full-width {
  grid-column: 1 / -1;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  flex-shrink: 0;
}

.card-header h3 {
  color: #ffffff;
  font-size: 1rem;
  font-weight: 600;
  margin: 0;
}

.monitoring-indicator {
  position: relative;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.ping-animation {
  position: relative;
  width: 24px;
  height: 24px;
}

.ping-dot {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 10px;
  height: 10px;
  background: #4ade80;
  border-radius: 50%;
  transform: translate(-50%, -50%);
  z-index: 2;
  box-shadow: 0 0 8px rgba(74, 222, 128, 0.6);
}

.ping-wave {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 10px;
  height: 10px;
  background: #4ade80;
  border-radius: 50%;
  transform: translate(-50%, -50%);
  animation: expanding-wave 1.5s cubic-bezier(0, 0, 0.2, 1) infinite;
  opacity: 0.6;
}

@keyframes expanding-wave {
  0% {
    transform: translate(-50%, -50%) scale(1);
    opacity: 0.8;
  }
  100% {
    transform: translate(-50%, -50%) scale(2.5);
    opacity: 0;
  }
}

.inactive-dot {
  width: 10px;
  height: 10px;
  background: #94a3b8;
  border-radius: 50%;
  opacity: 0.6;
}

.icon {
  color: #ff7c3d;
  flex-shrink: 0;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  filter: drop-shadow(0 2px 4px rgba(255, 124, 61, 0.3));
}

.icon.active {
  animation: bounce 1s infinite;
}

@keyframes bounce {
  0%, 100% { transform: translateY(0) scale(1); }
  50% { transform: translateY(-3px) scale(1.05); }
}

.card-content {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  flex: 1;
  justify-content: center;
}

.uptime-display {
  font-size: 1.8rem;
  font-weight: 700;
  color: #ff7c3d;
  text-align: center;
  text-shadow: 0 2px 8px rgba(255, 124, 61, 0.3);
  font-family: 'Courier New', monospace;
}

.monitor-status {
  font-size: 1.1rem;
  font-weight: 600;
  color: #ffffff;
  text-align: center;
  margin-bottom: 0.5rem;
}

.toggle-button {
  padding: 0.8rem;
  background: linear-gradient(135deg, #ff7c3d, #e55a2b);
  border: none;
  border-radius: 8px;
  color: #ffffff;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  text-transform: uppercase;
  letter-spacing: 1px;
  font-size: 0.85rem;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  width: 100%;
}

.toggle-button:hover {
  background: linear-gradient(135deg, #ff9966, #ff7c3d);
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4), 0 0 16px rgba(255, 124, 61, 0.2);
}

.toggle-button.stop {
  background: linear-gradient(135deg, #ef4444, #dc2626);
  color: #ffffff;
}

.toggle-button.stop:hover {
  background: linear-gradient(135deg, #f87171, #ef4444);
}

.monitor-path {
  color: #94a3b8;
  font-family: 'Courier New', monospace;
  font-size: 0.9rem;
  word-break: break-all;
  background: #2a3441;
  padding: 0.8rem;
  border-radius: 8px;
  border: 1px solid #3a4553;
}

.paths-container {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.monitor-path {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  padding: 0.75rem;
  background: #2a3441;
  border-radius: 8px;
  border: 1px solid #3a4553;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.monitor-path:hover {
  border-color: #4a5563;
  background: #364152;
}

.path-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8rem;
  font-weight: 600;
  color: #ffffff;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.path-value {
  color: #94a3b8;
  font-family: 'Courier New', monospace;
  font-size: 0.85rem;
  word-break: break-all;
  line-height: 1.4;
}

.path-icon {
  width: 18px;
  height: 18px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: bold;
  font-size: 0.7rem;
  flex-shrink: 0;
}

.medaltu-icon {
  background: #ff7c3d;
}

.nvidia-icon {
  background: #4ade80;
}

.custom-icon {
  background: #3b82f6;
}

.medaltu-path {
  border-color: rgba(255, 124, 61, 0.3);
}

.nvidia-path {
  border-color: rgba(74, 222, 128, 0.3);
}

.custom-path {
  border-color: rgba(59, 130, 246, 0.3);
}

.no-paths {
  border-color: rgba(148, 163, 184, 0.3);
  text-align: center;
}

.no-paths .path-value {
  color: #64748b;
  font-style: italic;
}

.clips-stats {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
  width: 100%;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem 0.8rem;
  background: #2a3441;
  border-radius: 12px;
  border: 1px solid #3a4553;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  min-height: fit-content;
}

.stat-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
  border-color: #4a5563;
  background: #364152;
}

.stat-label {
  color: #94a3b8;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  text-align: center;
  line-height: 1.2;
}

.stat-value {
  color: #ff7c3d;
  font-size: 2rem;
  font-weight: 800;
  font-family: 'Courier New', monospace;
  text-shadow: 0 2px 8px rgba(255, 124, 61, 0.3);
}

.loading-spinner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.5rem;
  padding: 3rem;
  flex: 1;
  justify-content: center;
}

.spinner {
  position: relative;
  width: 48px;
  height: 48px;
}

.spinner::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 16px;
  height: 16px;
  background: #ff7c3d;
  border-radius: 50%;
  transform: translate(-50%, -50%);
  z-index: 2;
  box-shadow: 0 0 12px rgba(255, 124, 61, 0.6);
}

.spinner::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 16px;
  height: 16px;
  background: #ff7c3d;
  border-radius: 50%;
  transform: translate(-50%, -50%);
  animation: spinnerPing 1.5s cubic-bezier(0, 0, 0.2, 1) infinite;
  opacity: 0.6;
}

@keyframes spinnerPing {
  0% {
    transform: translate(-50%, -50%) scale(1);
    opacity: 0.8;
  }
  75%, 100% {
    transform: translate(-50%, -50%) scale(3);
    opacity: 0;
  }
}
</style>
