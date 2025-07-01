<template>
  <transition name="modal" appear>
    <div v-if="isVisible" class="progress-overlay">
      <div class="progress-modal">
        <div class="progress-header">
          <h3>{{ title }}</h3>
          <button v-if="canClose" @click="$emit('close')" class="close-btn">&times;</button>
        </div>
        
        <div class="progress-body">
          <div class="progress-bar-container">
            <div class="progress-bar">
              <div 
                class="progress-fill" 
                :style="{ width: progressPercentage + '%' }"
                :class="{ 'error': hasError, 'complete': isComplete }"
              >
                <div class="progress-shine"></div>
              </div>
            </div>
            <div class="progress-text">{{ progressPercentage.toFixed(0) }}%</div>
          </div>
          
          <div class="progress-message">
            {{ currentMessage }}
          </div>
          
          <div v-if="hasError" class="error-details">
            {{ errorMessage }}
          </div>
          
          <div v-if="stage === 'compression'" class="compression-details">
            <div class="stage-indicator">
              <span class="stage-dot" :class="{ active: stage === 'compression' }"></span>
              Compressing video for optimal size and quality...
            </div>
          </div>
          
          <div class="button-group" v-if="isComplete || hasError">
            <button @click="$emit('close')" class="btn-primary">
              {{ hasError ? 'Close' : 'Done' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'

const props = defineProps({
  isVisible: {
    type: Boolean,
    default: false
  },
  title: {
    type: String,
    default: 'Processing...'
  },
  progress: {
    type: Number,
    default: 0
  },
  stage: {
    type: String,
    default: ''
  },
  message: {
    type: String,
    default: ''
  },
  error: {
    type: String,
    default: ''
  },
  isComplete: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['close'])

const progressPercentage = computed(() => {
  // Ensure smooth transition and proper completion
  const percentage = Math.max(0, Math.min(100, props.progress * 100))
  return props.isComplete ? 100 : percentage
})

const hasError = computed(() => props.error !== '')
const canClose = computed(() => props.isComplete || hasError.value)
const currentMessage = computed(() => {
  if (hasError.value) return 'Error occurred during processing'
  if (props.isComplete) return 'Process completed successfully!'
  return props.message || 'Processing...'
})
const errorMessage = computed(() => props.error)

// Show modal with delay to prevent jump scare
const showModal = ref(false)

watch(() => props.isVisible, async (newVal) => {
  if (newVal) {
    // Small delay before showing to prevent jump scare
    await nextTick()
    setTimeout(() => {
      showModal.value = true
    }, 100)
  } else {
    showModal.value = false
  }
})

// Auto-close after success with longer delay
watch(() => props.isComplete, (newVal) => {
  if (newVal) {
    setTimeout(() => {
      emit('close')
    }, 3000) // Auto-close after 3 seconds
  }
})
</script>

<style scoped>
/* Modal transition animations */
.modal-enter-active {
  transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1);
}

.modal-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.modal-enter-from {
  opacity: 0;
  transform: scale(0.8) translateY(-20px);
}

.modal-leave-to {
  opacity: 0;
  transform: scale(0.9) translateY(10px);
}

.progress-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(10, 13, 20, 0.85);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  backdrop-filter: blur(8px);
  animation: overlayFadeIn 0.4s cubic-bezier(0.175, 0.885, 0.32, 1);
  padding: 16px;
  box-sizing: border-box;
}

@keyframes overlayFadeIn {
  from {
    opacity: 0;
    backdrop-filter: blur(0px);
  }
  to {
    opacity: 1;
    backdrop-filter: blur(8px);
  }
}

.progress-modal {
  background: var(--bg-cards, #1e2530);
  border: 1px solid var(--border-default, #3a4553);
  border-radius: 16px;
  box-shadow: var(--shadow-xl, 0 16px 48px rgba(0, 0, 0, 0.7));
  color: var(--text-primary, #ffffff);
  width: 100%;
  max-width: 480px;
  min-width: 320px;
  max-height: 90vh;
  padding: 0;
  position: relative;
  overflow: hidden;
  transform: translateY(0);
  transition: all 0.3s cubic-bezier(0.175, 0.885, 0.32, 1);
}

.progress-modal::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, var(--primary-color, #ff7c3d), var(--primary-light, #ff9966));
  z-index: 1;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24px 28px 16px;
  border-bottom: 1px solid var(--border-default, #3a4553);
  background: var(--bg-elements, #2a3441);
}

.progress-header h3 {
  margin: 0;
  font-size: 1.3rem;
  font-weight: 600;
  color: var(--text-primary, #ffffff);
}

.close-btn {
  background: var(--bg-interactive, #364152);
  border: 1px solid var(--border-light, #4a5563);
  color: var(--text-secondary, #e2e8f0);
  font-size: 20px;
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  transition: var(--transition-fast);
}

.close-btn:hover {
  background: var(--bg-cards, #1e2530);
  border-color: var(--border-accent, rgba(255, 124, 61, 0.5));
  color: var(--text-primary, #ffffff);
}

.progress-body {
  padding: 24px 28px 28px;
  background: var(--bg-cards, #1e2530);
}

.progress-bar-container {
  margin-bottom: 20px;
  position: relative;
}

.progress-bar {
  width: 100%;
  height: 12px;
  background: var(--bg-elements, #2a3441);
  border: 1px solid var(--border-default, #3a4553);
  border-radius: 8px;
  overflow: hidden;
  position: relative;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--primary-color, #ff7c3d), var(--primary-light, #ff9966));
  border-radius: 7px;
  transition: width 0.8s cubic-bezier(0.25, 0.46, 0.45, 0.94);
  position: relative;
  overflow: hidden;
  min-width: 0;
}

.progress-shine {
  position: absolute;
  top: 0;
  left: -50%;
  width: 50%;
  height: 100%;
  background: linear-gradient(90deg, 
    transparent, 
    rgba(255, 255, 255, 0.4) 50%, 
    transparent
  );
  animation: shine 3s infinite cubic-bezier(0.25, 0.46, 0.45, 0.94);
}

@keyframes shine {
  0% { 
    left: -50%; 
    opacity: 0; 
  }
  20% { 
    opacity: 1; 
  }
  80% { 
    opacity: 1; 
  }
  100% { 
    left: 100%; 
    opacity: 0; 
  }
}

.progress-fill.error {
  background: linear-gradient(90deg, var(--error-color, #ff6b6b), #ff8a8a);
}

.progress-fill.error .progress-shine {
  display: none;
}

.progress-fill.complete {
  background: linear-gradient(90deg, var(--success-color, #4ade80), #65e695);
  width: 100% !important;
}

.progress-fill.complete .progress-shine {
  animation: completeShine 1s ease-out;
}

@keyframes completeShine {
  0% { 
    left: -50%; 
    opacity: 0; 
  }
  50% { 
    opacity: 1; 
  }
  100% { 
    left: 100%; 
    opacity: 0; 
  }
}

.progress-text {
  text-align: center;
  font-size: 0.9rem;
  margin-top: 12px;
  font-weight: 600;
  color: var(--text-secondary, #e2e8f0);
  font-family: 'Segoe UI', monospace;
}

.progress-message {
  text-align: center;
  margin-bottom: 16px;
  font-size: 1rem;
  line-height: 1.5;
  min-height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary, #e2e8f0);
  padding: 0 12px;
}

.error-details {
  background: rgba(255, 107, 107, 0.1);
  border: 1px solid rgba(255, 107, 107, 0.3);
  border-radius: 8px;
  padding: 16px;
  margin-top: 16px;
  font-size: 0.9rem;
  line-height: 1.5;
  word-break: break-word;
  color: var(--error-color, #ff6b6b);
}

.compression-details {
  margin: 20px 0;
  padding: 16px;
  background: var(--bg-elements, #2a3441);
  border: 1px solid var(--border-default, #3a4553);
  border-radius: 8px;
  font-size: 0.9rem;
}

.stage-indicator {
  display: flex;
  align-items: center;
  gap: 12px;
  color: var(--text-muted, #94a3b8);
}

.stage-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--border-light, #4a5563);
  transition: all 0.3s ease;
  flex-shrink: 0;
}

.stage-dot.active {
  background: var(--primary-color, #ff7c3d);
  box-shadow: 0 0 12px rgba(255, 124, 61, 0.6);
  animation: pulse 2.5s infinite;
}

@keyframes pulse {
  0%, 100% { 
    transform: scale(1); 
    opacity: 1; 
  }
  50% { 
    transform: scale(1.3); 
    opacity: 0.8; 
  }
}

.button-group {
  display: flex;
  justify-content: center;
  gap: 16px;
  margin-top: 24px;
}

.btn-primary {
  background: var(--primary-color, #ff7c3d);
  color: white;
  border: none;
  padding: 12px 28px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 0.95rem;
  font-weight: 600;
  transition: var(--transition-fast);
  box-shadow: var(--shadow-small, 0 2px 8px rgba(0, 0, 0, 0.4));
}

.btn-primary:hover {
  background: var(--primary-dark, #e55a2b);
  transform: translateY(-1px);
  box-shadow: var(--shadow-medium, 0 4px 16px rgba(0, 0, 0, 0.5));
}

.btn-primary:active {
  transform: translateY(0);
}

/* Responsive design for smaller screens */
@media (max-width: 768px) {
  .progress-modal {
    max-width: 95vw;
    min-width: 280px;
    margin: 0 8px;
  }
  
  .progress-header {
    padding: 20px 20px 12px;
  }
  
  .progress-header h3 {
    font-size: 1.1rem;
  }
  
  .progress-body {
    padding: 20px 20px 24px;
  }
  
  .progress-message {
    font-size: 0.9rem;
    min-height: 40px;
  }
}

@media (max-width: 480px) {
  .progress-modal {
    max-width: 96vw;
    min-width: 260px;
  }
  
  .progress-header {
    padding: 16px 16px 8px;
  }
  
  .progress-body {
    padding: 16px 16px 20px;
  }
}
</style>
