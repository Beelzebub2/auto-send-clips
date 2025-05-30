<template>
  <div class="toast-container">
    <TransitionGroup name="toast" tag="div">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        :class="['toast', `toast-${toast.type}`]"
        @click="removeToast(toast.id)"
      >
        <div class="toast-icon">
          <component :is="getToastIcon(toast.type)" :size="20" />
        </div>
        <div class="toast-content">
          <div class="toast-title">{{ toast.title }}</div>
          <div class="toast-message">{{ toast.message }}</div>
        </div>
        <button class="toast-close" @click.stop="removeToast(toast.id)">
          <X :size="16" />
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { X, CheckCircle, AlertCircle, Info, AlertTriangle } from 'lucide-vue-next'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

const toasts = ref([])
let toastIdCounter = 0

const getToastIcon = (type) => {
  const icons = {
    success: CheckCircle,
    error: AlertCircle,
    warning: AlertTriangle,
    info: Info
  }
  return icons[type] || Info
}

const addToast = (toast) => {
  const id = ++toastIdCounter
  const newToast = {
    id,
    type: toast.type || 'info',
    title: toast.title || 'Notification',
    message: toast.message || '',
    duration: toast.duration || 5000
  }
  
  toasts.value.push(newToast)
  
  // Auto remove after duration
  setTimeout(() => {
    removeToast(id)
  }, newToast.duration)
}

const removeToast = (id) => {
  const index = toasts.value.findIndex(toast => toast.id === id)
  if (index > -1) {
    toasts.value.splice(index, 1)
  }
}

// Listen for various events and show appropriate toasts
onMounted(() => {
  // Configuration saved
  EventsOn('config-saved', () => {
    addToast({
      type: 'success',
      title: 'Settings Saved',
      message: 'Your configuration has been saved successfully!',
      duration: 3000
    })
  })
  
  // Configuration error
  EventsOn('config-error', (data) => {
    addToast({
      type: 'error',
      title: 'Configuration Error',
      message: data.message || 'Failed to save configuration',
      duration: 5000
    })
  })
  
  // Monitoring started
  EventsOn('monitoring-started', () => {
    addToast({
      type: 'success',
      title: 'Monitoring Started',
      message: 'File monitoring is now active',
      duration: 3000
    })
  })
  
  // Monitoring stopped
  EventsOn('monitoring-stopped', () => {
    addToast({
      type: 'info',
      title: 'Monitoring Stopped',
      message: 'File monitoring has been stopped',
      duration: 3000
    })
  })
  
  // Video detected
  EventsOn('video-detected', (data) => {
    addToast({
      type: 'info',
      title: 'Video Detected',
      message: `New video file: ${data.fileName}`,
      duration: 4000
    })
  })
  
  // Video sent
  EventsOn('video-sent', (data) => {
    addToast({
      type: 'success',
      title: 'Video Sent',
      message: `Successfully sent: ${data.fileName}`,
      duration: 4000
    })
  })
  
  // Webhook test
  EventsOn('webhook-test-success', () => {
    addToast({
      type: 'success',
      title: 'Webhook Test',
      message: 'Webhook test successful!',
      duration: 3000
    })
  })
  
  EventsOn('webhook-test-error', (data) => {
    addToast({
      type: 'error',
      title: 'Webhook Test Failed',
      message: data.message || 'Webhook test failed',
      duration: 5000
    })
  })
})

onUnmounted(() => {
  EventsOff('config-saved')
  EventsOff('config-error')
  EventsOff('monitoring-started')
  EventsOff('monitoring-stopped')
  EventsOff('video-detected')
  EventsOff('video-sent')
  EventsOff('webhook-test-success')
  EventsOff('webhook-test-error')
})
</script>

<style scoped>
.toast-container {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 10000;
  display: flex;
  flex-direction: column;
  gap: 12px;
  pointer-events: none;
}

.toast {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  min-width: 300px;
  max-width: 400px;
  padding: 16px;
  background: rgba(20, 20, 20, 0.95);
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(20px);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
  cursor: pointer;
  pointer-events: auto;
  transition: all 0.3s ease;
}

.toast:hover {
  transform: translateX(-4px);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.6);
}

.toast-success {
  border-left: 4px solid #10b981;
}

.toast-error {
  border-left: 4px solid #ef4444;
}

.toast-warning {
  border-left: 4px solid #f59e0b;
}

.toast-info {
  border-left: 4px solid #3b82f6;
}

.toast-icon {
  flex-shrink: 0;
  margin-top: 2px;
}

.toast-success .toast-icon {
  color: #10b981;
}

.toast-error .toast-icon {
  color: #ef4444;
}

.toast-warning .toast-icon {
  color: #f59e0b;
}

.toast-info .toast-icon {
  color: #3b82f6;
}

.toast-content {
  flex: 1;
  min-width: 0;
}

.toast-title {
  font-weight: 600;
  font-size: 0.9rem;
  color: #ffffff;
  margin-bottom: 4px;
}

.toast-message {
  font-size: 0.8rem;
  color: rgba(255, 255, 255, 0.8);
  line-height: 1.4;
  word-break: break-word;
}

.toast-close {
  flex-shrink: 0;
  background: none;
  border: none;
  color: rgba(255, 255, 255, 0.6);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s ease;
  margin-top: -2px;
}

.toast-close:hover {
  color: #ffffff;
  background: rgba(255, 255, 255, 0.1);
}

/* Toast animations */
.toast-enter-active {
  transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

.toast-leave-active {
  transition: all 0.3s ease-in;
}

.toast-enter-from {
  transform: translateX(100%);
  opacity: 0;
}

.toast-leave-to {
  transform: translateX(100%);
  opacity: 0;
}

.toast-move {
  transition: transform 0.3s ease;
}

/* Responsive design */
@media (max-width: 768px) {
  .toast-container {
    top: 10px;
    right: 10px;
    left: 10px;
  }
  
  .toast {
    min-width: auto;
    max-width: none;
  }
}

@media (max-width: 480px) {
  .toast {
    padding: 12px;
    gap: 8px;
  }
  
  .toast-title {
    font-size: 0.85rem;
  }
  
  .toast-message {
    font-size: 0.75rem;
  }
}
</style>
