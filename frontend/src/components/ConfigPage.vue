<script setup>
import { ref, onMounted } from 'vue'
import { Settings, Globe, Folder, FolderOpen, TestTube, Save, ChevronDown, ChevronUp } from 'lucide-vue-next'
import { GetConfig, SaveConfig, UpdateMonitorPath, SelectFolder } from '../../wailsjs/go/main/App'

const config = ref({
  webhookURL: '',
  monitorPath: '',
  maxFileSize: 10485760, // 10MB in bytes
  checkInterval: 2
})

const isLoading = ref(true)
const isSaving = ref(false)
const error = ref('')
const success = ref('')
const showAdvanced = ref(false)

const loadConfig = async () => {
  try {
    isLoading.value = true
    error.value = ''
    const appConfig = await GetConfig()
    config.value = {
      webhookURL: appConfig.webhook_url || '',
      monitorPath: appConfig.monitor_path || '',
      maxFileSize: appConfig.max_file_size || 10485760,
      checkInterval: appConfig.check_interval || 2
    }
  } catch (err) {
    error.value = 'Failed to load configuration: ' + err.message
  } finally {
    isLoading.value = false
  }
}

const saveConfig = async () => {
  try {
    isSaving.value = true
    error.value = ''
    success.value = ''
    
    const configToSave = {
      webhook_url: config.value.webhookURL,
      monitor_path: config.value.monitorPath,
      max_file_size: config.value.maxFileSize,
      check_interval: config.value.checkInterval
    }
    
    await SaveConfig(configToSave)
    success.value = 'Configuration saved successfully!'
    
    // Clear success message after 3 seconds
    setTimeout(() => {
      success.value = ''
    }, 3000)
  } catch (err) {
    error.value = 'Failed to save configuration: ' + err.message
  } finally {
    isSaving.value = false
  }
}

const selectFolder = async () => {
  try {
    const selectedPath = await SelectFolder()
    if (selectedPath) {
      config.value.monitorPath = selectedPath
    }
  } catch (err) {
    if (err && err.message) {
      error.value = 'Failed to select folder: ' + err.message
    }
    // If no folder was selected (cancel), do nothing
  }
}

const updateMonitorPath = async () => {
  try {
    await UpdateMonitorPath(config.value.monitorPath)
    success.value = 'Monitor path updated and watcher restarted!'
    setTimeout(() => {
      success.value = ''
    }, 3000)
  } catch (err) {
    error.value = 'Failed to update monitor path: ' + err.message
  }
}

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const testWebhook = async () => {
  if (!config.value.webhookURL) {
    error.value = 'Please enter a webhook URL first'
    return
  }
  
  try {
    // Simple test - you might want to implement this in Go
    const response = await fetch(config.value.webhookURL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        content: 'Test message from AutoClipSend ✅'
      })
    })
    
    if (response.ok) {
      success.value = 'Webhook test successful!'
    } else {
      error.value = 'Webhook test failed: ' + response.status
    }
  } catch (err) {
    error.value = 'Webhook test failed: ' + err.message
  }
}

onMounted(() => {
  loadConfig()
})
</script>

<template>  <div class="config-page">
    <div class="config-header">
      <h2>Configuration Settings</h2>
    </div>

    <div class="error-message" v-if="error">
      {{ error }}
    </div>

    <div class="success-message" v-if="success">
      {{ success }}
    </div>

    <div class="config-form" v-if="!isLoading">
      <!-- Discord Configuration -->
      <div class="config-section">
        <h3>
          <Globe :size="18" />
          Discord Integration
        </h3>
        <div class="form-group">
          <label for="webhookUrl">Discord Webhook URL</label>
          <div class="input-group">
            <input
              id="webhookUrl"
              v-model="config.webhookURL"
              type="url"
              placeholder="https://discord.com/api/webhooks/..."
              class="form-input"
            />
            <button @click="testWebhook" class="test-button" :disabled="!config.webhookURL">
              <TestTube :size="16" />
              Test
            </button>
          </div>
          <p class="form-help">
            Get your webhook URL from Discord: Server Settings → Integrations → Webhooks
          </p>
        </div>
      </div>

      <!-- Monitor Configuration -->
      <div class="config-section">
        <h3>
          <Folder :size="18" />
          File Monitoring
        </h3>
        <div class="form-group">
          <label for="monitorPath">Monitor Path</label>
          <div class="input-group">
            <input
              id="monitorPath"
              v-model="config.monitorPath"
              type="text"
              placeholder="Select a folder to monitor"
              class="form-input"
              readonly
            />
            <button @click="selectFolder" class="folder-button">
              <FolderOpen :size="16" />
              Browse
            </button>
            <button @click="updateMonitorPath" class="update-button" :disabled="!config.monitorPath">
              <Save :size="16" />
              Update
            </button>
          </div>
          <p class="form-help">
            The folder path to monitor for new video files
          </p>
        </div>
      </div>

      <!-- Advanced Settings -->
      <div class="config-section" v-if="showAdvanced">
        <h3>
          <Settings :size="18" />
          Advanced Settings
        </h3>
        <div class="form-row">
          <div class="form-group">
            <label for="maxFileSize">Max File Size (MB)</label>
            <input
              id="maxFileSize"
              v-model.number="config.maxFileSize"
              type="number"
              min="1"
              max="100"
              class="form-input"
              @input="config.maxFileSize = $event.target.value * 1024 * 1024"
            />
            <p class="form-help">
              Current: {{ formatFileSize(config.maxFileSize) }}
            </p>
          </div>
          
          <div class="form-group">
            <label for="checkInterval">Check Interval (seconds)</label>
            <input
              id="checkInterval"
              v-model.number="config.checkInterval"
              type="number"
              min="1"
              max="60"
              class="form-input"
            />
            <p class="form-help">
              How long to wait before processing new files
            </p>
          </div>
        </div>
      </div>

      <!-- Save Button -->
      <div class="form-actions">
        <button 
          @click="saveConfig" 
          class="save-button" 
          :disabled="isSaving"
        >
          <Save :size="18" />
          <span v-if="isSaving">Saving...</span>
          <span v-else">Save Configuration</span>
        </button>
      </div>

      <!-- Advanced Toggle at Bottom -->
      <div class="advanced-toggle-container">
        <button @click="showAdvanced = !showAdvanced" class="toggle-advanced">
          <component :is="showAdvanced ? ChevronUp : ChevronDown" :size="16" />
          {{ showAdvanced ? 'Hide Advanced' : 'Show Advanced' }}
        </button>
      </div>
    </div>

    <div class="loading-spinner" v-if="isLoading">
      <div class="spinner"></div>
      <span>Loading configuration...</span>
    </div>
  </div>
</template>

<style scoped>
.config-page {
  height: 100vh;
  overflow-y: auto;
  padding: 1rem;
  display: flex;
  flex-direction: column;
}

.config-header {
  display: flex;
  justify-content: center;
  align-items: center;
  margin-bottom: 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid rgba(255, 140, 0, 0.3);
  flex-shrink: 0;
}

.config-header h2 {
  color: #ff8c00;
  font-size: 1.2rem;
  font-weight: 600;
  margin: 0;
}

.advanced-toggle-container {
  display: flex;
  justify-content: center;
  padding-top: 1rem;
  margin-top: 1rem;
  border-top: 1px solid rgba(255, 140, 0, 0.2);
}

.toggle-advanced {
  padding: 0.4rem 0.8rem;
  background: transparent;
  border: 1px solid rgba(255, 140, 0, 0.3);
  color: #ff8c00;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.3s ease;
  font-size: 0.8rem;
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.toggle-advanced:hover {
  background: rgba(255, 140, 0, 0.1);
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

.success-message {
  background: rgba(0, 255, 0, 0.1);
  border: 1px solid rgba(0, 255, 0, 0.3);
  color: #4ade80;
  padding: 0.75rem;
  border-radius: 8px;
  margin-bottom: 1rem;
  font-size: 0.9rem;
  flex-shrink: 0;
}

.config-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  flex: 1;
  min-height: 0;
}

.config-section {
  background: rgba(0, 0, 0, 0.3);
  border-radius: 8px;
  padding: 1rem;
  border: 1px solid rgba(255, 140, 0, 0.2);
  backdrop-filter: blur(10px);
}

.config-section h3 {
  color: #ffffff;
  font-size: 1rem;
  font-weight: 600;
  margin-bottom: 1rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.4rem;
  color: #ffffff;
  font-weight: 500;
  font-size: 0.9rem;
}

.form-input {
  width: 100%;
  padding: 0.6rem;
  background: rgba(0, 0, 0, 0.4);
  border: 1px solid rgba(255, 140, 0, 0.3);
  border-radius: 6px;
  color: #ffffff;
  font-size: 0.85rem;
  transition: all 0.3s ease;
}

.form-input:focus {
  outline: none;
  border-color: #ff8c00;
  box-shadow: 0 0 0 2px rgba(255, 140, 0, 0.2);
}

.form-input::placeholder {
  color: rgba(255, 255, 255, 0.4);
}

.input-group {
  display: flex;
  gap: 0.4rem;
}

.input-group .form-input {
  flex: 1;
}

.test-button, .update-button, .folder-button {
  padding: 0.6rem 0.8rem;
  background: #ff8c00;
  border: none;
  border-radius: 6px;
  color: #1a1a1a;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  white-space: nowrap;
  display: flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.8rem;
}

.test-button:hover, .update-button:hover, .folder-button:hover {
  background: #e67e22;
}

.test-button:disabled, .update-button:disabled {
  background: rgba(255, 140, 0, 0.3);
  cursor: not-allowed;
}

.folder-button {
  background: #2196f3;
}

.folder-button:hover {
  background: #1976d2;
}

.form-help {
  margin-top: 0.4rem;
  font-size: 0.75rem;
  color: rgba(255, 255, 255, 0.6);
  line-height: 1.3;
}

.form-actions {
  display: flex;
  justify-content: center;
  padding-top: 1rem;
  margin-top: auto;
  flex-shrink: 0;
}

.save-button {
  padding: 0.8rem 1.5rem;
  background: linear-gradient(135deg, #ff8c00 0%, #e67e22 100%);
  border: none;
  border-radius: 8px;
  color: #1a1a1a;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  min-width: 180px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.save-button:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(255, 140, 0, 0.3);
}

.save-button:disabled {
  background: rgba(255, 140, 0, 0.3);
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
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

/* Scrollbar styling */
.config-page::-webkit-scrollbar {
  width: 6px;
}

.config-page::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.1);
}

.config-page::-webkit-scrollbar-thumb {
  background: rgba(255, 140, 0, 0.3);
  border-radius: 3px;
}

.config-page::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 140, 0, 0.5);
}
</style>
