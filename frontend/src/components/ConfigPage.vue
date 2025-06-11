<script setup>
import { ref, onMounted, watch } from 'vue'
import { Settings, Globe, Folder, FolderOpen, TestTube, Save } from 'lucide-vue-next'
import { GetConfig, SaveConfig, UpdateMonitorPath, SelectFolder, SetWindowsStartup } from '../../wailsjs/go/main/App'

const config = ref({
  webhookURL: '',
  monitorPath: '',
  maxFileSize: 20, // Store in MB directly
  checkInterval: 2,
  startupInitialization: true,
  windowsStartup: false,  recursiveMonitoring: false
})

const isSaving = ref(false)
const error = ref('')
// Use a completely separate reactive property to control animations
const animateNow = ref(false)

const loadConfig = async () => {
  try {
    animateNow.value = false 
    error.value = ''
    const appConfig = await GetConfig()
    config.value = {
      webhookURL: appConfig.webhook_url || '',
      monitorPath: appConfig.monitor_path || '',
      maxFileSize: appConfig.max_file_size || 10, // Backend stores in MB now
      checkInterval: appConfig.check_interval || 2,
      startupInitialization: appConfig.startup_initialization !== undefined ? appConfig.startup_initialization : true,
      windowsStartup: appConfig.windows_startup !== undefined ? appConfig.windows_startup : false,
      recursiveMonitoring: appConfig.recursive_monitoring !== undefined ? appConfig.recursive_monitoring : false
    }
  } catch (err) {
    error.value = 'Failed to load configuration: ' + err.message
  } finally {
    // Force browser to do a paint cycle before triggering animations
    document.body.offsetHeight; // Force a reflow
      // Then trigger animations after a very brief delay
    setTimeout(() => {
      animateNow.value = true
      console.log('Animations triggered at', new Date())
    }, 50)
  }
}

const saveConfig = async () => {
  try {    isSaving.value = true
    error.value = ''
    const configToSave = {
      webhook_url: config.value.webhookURL,
      monitor_path: config.value.monitorPath,
      max_file_size: config.value.maxFileSize,
      check_interval: config.value.checkInterval,
      startup_initialization: config.value.startupInitialization,
      windows_startup: config.value.windowsStartup,
      recursive_monitoring: config.value.recursiveMonitoring
    }
    
    await SaveConfig(configToSave)
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
  } catch (err) {
    error.value = 'Failed to update monitor path: ' + err.message
  }
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
      // Simple success feedback
    } else {
      error.value = 'Webhook test failed: HTTP ' + response.status
    }
  } catch (err) {
    error.value = 'Webhook test failed: ' + err.message
  }
}

onMounted(() => {
  // Reset animations flag first
  animateNow.value = false
  
  // Load config data with a small delay to ensure all reactivity is established
  setTimeout(() => {
    loadConfig()
  }, 10)
})

// Watch for Windows startup setting changes and immediately apply them
watch(() => config.value.windowsStartup, async (newValue) => {
  try {
    await SetWindowsStartup(newValue)
  } catch (err) {
    error.value = 'Failed to update Windows startup setting: ' + err.message
    // Revert the toggle if the backend call failed
    config.value.windowsStartup = !newValue
  }
})

// Auto-save watchers for immediate configuration updates
let saveTimeout = null

const debouncedSave = () => {
  if (saveTimeout) {
    clearTimeout(saveTimeout)
  }
  saveTimeout = setTimeout(async () => {
    await saveConfig()
  }, 500) // 500ms debounce
}

watch(() => config.value.webhookURL, debouncedSave, { deep: true })

watch(() => config.value.monitorPath, async (newPath, oldPath) => {
  // Only auto-save and update monitor path if the path actually changed and is not empty
  if (newPath && newPath !== oldPath) {
    await saveConfig()
    await updateMonitorPath()
  }
}, { deep: true })

watch(() => config.value.maxFileSize, debouncedSave, { deep: true })
watch(() => config.value.checkInterval, debouncedSave, { deep: true })
watch(() => config.value.startupInitialization, debouncedSave, { deep: true })

// Watch for recursive monitoring changes and immediately restart the watcher
watch(() => config.value.recursiveMonitoring, async () => {
  await saveConfig()
  // Restart the file watcher with the new recursive setting
  if (config.value.monitorPath) {
    await updateMonitorPath()
  }
}, { deep: true })
</script>

<template>  <div class="config-page">
    <transition name="fade-slide-up" :duration="{ enter: 800, leave: 600 }">
      <div class="config-header" v-if="animateNow">
        <h2>Configuration Settings</h2>
      </div>
    </transition>

    <transition name="fade-slide-up" :duration="{ enter: 800, leave: 600 }">
      <div class="error-message" v-if="error && animateNow">
        {{ error }}
      </div>
    </transition>    <transition name="fade" :duration="{ enter: 800, leave: 600 }">
      <div class="config-form" v-if="animateNow">
        <div class="config-grid">
          <!-- Left Column - Main Settings -->
          <div class="config-column main-settings">            <!-- Discord Configuration -->
            <transition name="fade-slide-up" :duration="{ enter: 800, leave: 600 }" appear>
              <div class="config-section primary" v-if="animateNow">
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
            </transition>
              <!-- File Monitoring -->
            <transition name="fade-slide-up" :duration="{ enter: 800, leave: 600 }" appear>
              <div class="config-section primary" v-if="animateNow">
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
                  </div>
                  <p class="form-help">
                    The folder path to monitor for new video files
                  </p>
                </div>
                
                <div class="form-group">
                  <label for="recursiveMonitoring">Watch subfolders recursively</label>
                  <div class="toggle-container">
                    <label class="toggle-switch">
                      <input
                        id="recursiveMonitoring"
                        v-model="config.recursiveMonitoring"
                        type="checkbox"
                        class="toggle-input"
                      />
                      <span class="toggle-slider"></span>
                    </label>
                    <span class="toggle-label">{{ config.recursiveMonitoring ? 'Enabled' : 'Disabled' }}</span>
                  </div>
                  <p class="form-help">
                    Monitor all subfolders within the selected path for new video files
                  </p>
                </div>
              </div>
            </transition>
          </div>
          
          <!-- Right Column - Advanced Settings -->
          <div class="config-column monitoring-settings">            <transition name="fade-slide-up" :duration="{ enter: 800, leave: 600 }" appear>
              <div class="config-section secondary" v-if="animateNow">
                <h3>
                  <Settings :size="18" />
                  Advanced Settings
                </h3>
                <div class="form-group">
                  <label for="maxFileSize">Max File Size (MB)</label>
                  <input
                    id="maxFileSize"
                    v-model.number="config.maxFileSize"
                    type="number"
                    min="1"
                    max="8000"
                    class="form-input"
                  />
                  <p class="form-help">
                    Maximum file size in megabytes (MB)
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
                    How often to check for new files
                  </p>
                </div>
                
                <div class="form-group">
                  <label for="startupInitialization">Start monitoring on startup</label>
                  <div class="toggle-container">
                    <label class="toggle-switch">
                      <input
                        id="startupInitialization"
                        v-model="config.startupInitialization"
                        type="checkbox"
                        class="toggle-input"
                      />
                      <span class="toggle-slider"></span>
                    </label>
                    <span class="toggle-label">{{ config.startupInitialization ? 'Enabled' : 'Disabled' }}</span>
                  </div>
                  <p class="form-help">
                    Automatically begin file monitoring when the application starts
                  </p>
                </div>
                
                <div class="form-group">
                  <label for="windowsStartup">Start with Windows</label>
                  <div class="toggle-container">
                    <label class="toggle-switch">
                      <input
                        id="windowsStartup"
                        v-model="config.windowsStartup"
                        type="checkbox"
                        class="toggle-input"
                      />
                      <span class="toggle-slider"></span>
                    </label>
                    <span class="toggle-label">{{ config.windowsStartup ? 'Enabled' : 'Disabled' }}</span>
                  </div>
                  <p class="form-help">
                    Automatically launch AutoClipSend when Windows starts
                  </p>
                </div>
              </div>
            </transition>
          </div>
        </div>
      </div>    </transition>
  </div>
</template>

<style scoped>
.config-page {
  height: 100vh;
  overflow: auto;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  background: var(--bg-darkest);
}

.config-header {
  display: flex;
  justify-content: center;
  align-items: center;
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--border-default);
  flex-shrink: 0;
}

.config-header h2 {
  color: var(--primary-color);
  font-size: 1.3rem;
  font-weight: 700;
  margin: 0;
  text-shadow: 0 2px 8px rgba(255, 124, 61, 0.3);
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

.config-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  flex: 1;
  min-height: 0;
}

.config-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1.5rem;
  flex: 1;
  align-content: start;
}

.config-column {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.main-settings {
  flex: 1;
}

.advanced-settings {
  min-width: 0;
}

.monitoring-settings {
  min-width: 0;
}

.config-section {
  background: var(--bg-cards);
  border-radius: 16px;
  padding: 1.5rem;
  border: 1px solid var(--border-default);
  backdrop-filter: blur(20px);
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.3);
  position: relative;
  overflow: hidden;
  transition: transform 0.5s cubic-bezier(0.19, 1, 0.22, 1),
              border-color 0.5s cubic-bezier(0.19, 1, 0.22, 1),
              box-shadow 0.5s cubic-bezier(0.19, 1, 0.22, 1),
              background 0.5s cubic-bezier(0.19, 1, 0.22, 1);
}

.config-section::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, var(--primary-color), var(--primary-light));
  opacity: 0;
  transition: all 0.5s cubic-bezier(0.19, 1, 0.22, 1);
}

.config-section:hover {
  border-color: var(--border-light);
  transform: translateY(-4px) scale(1.01);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4), 0 0 24px rgba(255, 124, 61, 0.1);
  background: var(--bg-elements);
  transition: transform 0.5s cubic-bezier(0.19, 1, 0.22, 1),
              border-color 0.5s cubic-bezier(0.19, 1, 0.22, 1),
              box-shadow 0.5s cubic-bezier(0.19, 1, 0.22, 1),
              background 0.5s cubic-bezier(0.19, 1, 0.22, 1);
}

.config-section:hover::before {
  opacity: 1;
}

.config-section.primary {
  border-color: var(--border-accent);
}

.config-section.secondary {
  border-color: rgba(138, 43, 226, 0.4);
  background: linear-gradient(135deg, var(--bg-cards), rgba(138, 43, 226, 0.05));
}

.config-section.secondary:hover {
  border-color: rgba(138, 43, 226, 0.6);
  box-shadow: 0 8px 32px rgba(138, 43, 226, 0.2);
}

.config-section h3 svg {
  color: var(--primary-color);
  filter: drop-shadow(0 2px 4px rgba(255, 124, 61, 0.3));
  transition: all 0.5s cubic-bezier(0.19, 1, 0.22, 1);
}

.config-section:hover h3 svg {
  animation: bounce 1.5s infinite cubic-bezier(0.19, 1, 0.22, 1);
}

@keyframes bounce {
  0%, 100% { transform: translateY(0) scale(1); }
  50% { transform: translateY(-3px) scale(1.05); }
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
  transition: all 0.5s cubic-bezier(0.19, 1, 0.22, 1);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.form-input:focus {
  outline: none;
  border-color: #ff8c00;
  box-shadow: 0 0 0 3px rgba(255, 140, 0, 0.2), 0 4px 8px rgba(0, 0, 0, 0.3);
  transform: translateY(-1px);
  background: rgba(0, 0, 0, 0.6);
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

.test-button, .update-button, .folder-button, .update-path-button {
  padding: 0.6rem 0.8rem;
  background: #ff8c00;
  border: none;
  border-radius: 6px;
  color: #1a1a1a;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.5s cubic-bezier(0.19, 1, 0.22, 1);
  white-space: nowrap;
  display: flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.8rem;
  filter: drop-shadow(0 2px 4px rgba(255, 124, 61, 0.3));
}

.test-button:hover, .update-button:hover, .folder-button:hover, .update-path-button:hover {
  background: #e67e22;
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4), 0 0 16px rgba(255, 124, 61, 0.2);
}

.test-button:disabled, .update-button:disabled, .update-path-button:disabled {
  background: rgba(255, 140, 0, 0.3);
  cursor: not-allowed;
  transform: none;
  filter: none;
}

.folder-button {
  background: #2196f3;
}

.folder-button:hover {
  background: #1976d2;
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4), 0 0 16px rgba(33, 150, 243, 0.2);
}

.update-path-button {
  width: 100%;
  margin-top: 0.5rem;
  justify-content: center;
  background: #4CAF50;
}

.update-path-button:hover {
  background: #45a049;
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4), 0 0 16px rgba(76, 175, 80, 0.2);
}

.form-help {
  margin-top: 0.4rem;
  font-size: 0.75rem;
  color: rgba(255, 255, 255, 0.6);
  line-height: 1.3;
}

.checkbox-label {
  display: flex;
  align-items: center;
  cursor: pointer;
  font-weight: 500;
  color: #ffffff;
  font-size: 0.9rem;
}

.form-checkbox {
  margin-right: 0.75rem;
  accent-color: #ff8c00;
  transform: scale(1.2);
}

.checkbox-text {
  user-select: none;
}

/* Toggle Switch Styles */
.toggle-container {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.toggle-switch {
  position: relative;
  display: inline-block;
  width: 50px;
  height: 24px;
  cursor: pointer;
}

.toggle-input {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-slider {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  transition: all 0.5s cubic-bezier(0.19, 1, 0.22, 1);
  border-radius: 12px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.toggle-slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 2px;
  bottom: 2px;
  background-color: #ffffff;
  transition: all 0.5s cubic-bezier(0.19, 1, 0.22, 1);
  border-radius: 50%;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  transform: scale(1);
}

.toggle-input:checked + .toggle-slider {
  background-color: #ff8c00;
  border-color: #ff8c00;
  box-shadow: 0 0 12px rgba(255, 140, 0, 0.4);
}

.toggle-input:checked + .toggle-slider:before {
  transform: translateX(26px) scale(1.1);
}

.toggle-slider:hover {
  box-shadow: 0 0 8px rgba(255, 140, 0, 0.3);
  transform: scale(1.02);
}

.toggle-label {
  font-size: 0.9rem;
  font-weight: 500;
  color: #ffffff;
  min-width: 60px;
}

/* Vue transition classes */
.fade-slide-up-enter-active {
  transition: all 0.8s cubic-bezier(0.19, 1, 0.22, 1);
}

.fade-slide-up-leave-active {
  transition: all 0.6s cubic-bezier(0.19, 1, 0.22, 1);
}

.fade-slide-up-enter-from, 
.fade-slide-up-leave-to {
  opacity: 0;
  transform: translateY(20px);
}

.fade-enter-active {
  transition: opacity 0.8s cubic-bezier(0.19, 1, 0.22, 1);
}

.fade-leave-active {
  transition: opacity 0.6s cubic-bezier(0.19, 1, 0.22, 1);
}

.fade-enter-from, 
.fade-leave-to {
  opacity: 0;
}

/* Entrance animations */
@keyframes fadeInUp {
  0% {
    opacity: 0 !important;
    transform: translateY(20px) !important;
  }
  100% {
    opacity: 1 !important;
    transform: translateY(0) !important;
  }
}

@keyframes cardEntrance {
  0% {
    opacity: 0 !important;
    transform: translateY(30px) scale(0.95) !important;
  }
  30% {
    opacity: 0.5 !important;
  }
  100% {
    opacity: 1 !important;
    transform: translateY(0) scale(1) !important;
  }
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

/* Responsive design */
@media (max-width: 900px) {
  .config-grid {
    grid-template-columns: 1fr;
    gap: 1rem;
  }
  
  .config-section {
    padding: 1rem;
  }
  
  .input-group {
    flex-direction: column;
    gap: 0.5rem;
  }
}

@media (max-width: 768px) {
  .config-page {
    padding: 1rem;
  }
  
  .config-grid {
    gap: 0.75rem;
  }
  
  .config-section {
    padding: 0.75rem;
  }
}
</style>
