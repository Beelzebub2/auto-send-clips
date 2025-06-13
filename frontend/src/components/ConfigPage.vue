<script setup>
import { ref, onMounted, watch } from 'vue'
import { Settings, Globe, Folder, FolderOpen, TestTube, Save, Download, ExternalLink, Info, RefreshCw } from 'lucide-vue-next'
import { GetConfig, SaveConfig, UpdateMonitorPath, SelectFolder, SetWindowsStartup, SetDesktopShortcut, GetVersionInfo, CheckForUpdates, OpenUpdateURL, GetMedalTVClipFolder, GetNVIDIACurrentDirectory } from '../../wailsjs/go/main/App'

const config = ref({
  webhookURL: '',
  monitorPath: '',
  maxFileSize: 20, // Store in MB directly
  checkInterval: 2,
  startupInitialization: true,
  windowsStartup: false,
  recursiveMonitoring: false,
  desktopShortcut: false,
  useMedalTVPath: false,
  useNVIDIAPath: false,
  useCustomPath: false
})

const isSaving = ref(false)
const error = ref('')
// Use a completely separate reactive property to control animations
const animateNow = ref(false)

// Store individual paths for preview
const medalTVPath = ref('')
const nvidiaPath = ref('')

// Version and update related variables
const versionInfo = ref({})
const updateInfo = ref(null)
const isCheckingUpdates = ref(false)
const showVersionDetails = ref(false)

const loadConfig = async () => {
  try {    animateNow.value = false 
    error.value = ''
    const appConfig = await GetConfig()
    config.value = {
      webhookURL: appConfig.webhook_url || '',
      monitorPath: appConfig.monitor_path || '',
      maxFileSize: appConfig.max_file_size || 10, // Backend stores in MB now
      checkInterval: appConfig.check_interval || 2,
      startupInitialization: appConfig.startup_initialization !== undefined ? appConfig.startup_initialization : true,
      windowsStartup: appConfig.windows_startup !== undefined ? appConfig.windows_startup : false,
      recursiveMonitoring: appConfig.recursive_monitoring !== undefined ? appConfig.recursive_monitoring : false,
      desktopShortcut: appConfig.desktop_shortcut !== undefined ? appConfig.desktop_shortcut : false,
      useMedalTVPath: appConfig.use_medaltv_path !== undefined ? appConfig.use_medaltv_path : false,
      useNVIDIAPath: appConfig.use_nvidia_path !== undefined ? appConfig.use_nvidia_path : false,
      useCustomPath: appConfig.use_custom_path !== undefined ? appConfig.use_custom_path : false
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

const saveConfig = async () => {  try {
    isSaving.value = true
    error.value = ''
    const configToSave = {
      webhook_url: config.value.webhookURL,
      monitor_path: config.value.monitorPath,
      max_file_size: config.value.maxFileSize,
      check_interval: config.value.checkInterval,
      startup_initialization: config.value.startupInitialization,
      windows_startup: config.value.windowsStartup,
      recursive_monitoring: config.value.recursiveMonitoring,
      desktop_shortcut: config.value.desktopShortcut,
      use_medaltv_path: config.value.useMedalTVPath,
      use_nvidia_path: config.value.useNVIDIAPath,
      use_custom_path: config.value.useCustomPath
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

const getMedalTVPath = async () => {
  try {
    const medalPath = await GetMedalTVClipFolder()
    if (medalPath) {
      medalTVPath.value = medalPath
      await saveConfig()
      // Use restartMonitoring instead of updateMonitorPath for multi-path support
      await restartMonitoring()
    }
  } catch (err) {
    error.value = 'Failed to get MedalTV path: ' + err.message
    // Revert checkbox if getting path failed
    config.value.useMedalTVPath = false
  }
}

const getNVIDIAPath = async () => {
  try {
    const nvPath = await GetNVIDIACurrentDirectory()
    if (nvPath) {
      nvidiaPath.value = nvPath
      await saveConfig()
      // Use restartMonitoring instead of updateMonitorPath for multi-path support
      await restartMonitoring()
    }
  } catch (err) {
    error.value = 'Failed to get NVIDIA path: ' + err.message
    // Revert checkbox if getting path failed
    config.value.useNVIDIAPath = false
  }
}

const updateMonitorPath = async () => {
  try {
    await UpdateMonitorPath(config.value.monitorPath)
  } catch (err) {
    error.value = 'Failed to update monitor path: ' + err.message
  }
}

const restartMonitoring = async () => {
  try {
    // Import the new restart function - we'll add this to the backend
    const { RestartMonitoring } = await import('../../wailsjs/go/main/App')
    await RestartMonitoring()
  } catch (err) {
    // Fall back to the old method if the new function isn't available yet
    try {
      const { StopMonitoring, StartMonitoring } = await import('../../wailsjs/go/main/App')
      await StopMonitoring()
      // Small delay to ensure clean shutdown
      await new Promise(resolve => setTimeout(resolve, 100))
      await StartMonitoring()
    } catch (fallbackErr) {
      error.value = 'Failed to restart monitoring: ' + fallbackErr.message
    }
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

// Watch for desktop shortcut setting changes and immediately apply them
watch(() => config.value.desktopShortcut, async (newValue) => {
  try {
    await SetDesktopShortcut(newValue)
  } catch (err) {
    error.value = 'Failed to update desktop shortcut setting: ' + err.message
    // Revert the toggle if the backend call failed
    config.value.desktopShortcut = !newValue
  }
})

// Watch for custom path toggle changes
watch(() => config.value.useCustomPath, async (newValue) => {
  if (!newValue) {
    // User disabled custom path mode - clear the monitor path
    config.value.monitorPath = ''
  }
  await saveConfig()
  await restartMonitoring()
}, { deep: true })

// Watch for MedalTV checkbox changes
watch(() => config.value.useMedalTVPath, async (newValue) => {
  if (newValue) {
    // User enabled MedalTV mode - get the path
    await getMedalTVPath()
  } else {
    // User disabled MedalTV mode - clear the path and restart monitoring
    medalTVPath.value = ''
    await saveConfig()
    await restartMonitoring()
  }
}, { deep: true })

// Watch for NVIDIA checkbox changes
watch(() => config.value.useNVIDIAPath, async (newValue) => {
  if (newValue) {
    // User enabled NVIDIA mode - get the path
    await getNVIDIAPath()
  } else {
    // User disabled NVIDIA mode - clear the path and restart monitoring
    nvidiaPath.value = ''
    await saveConfig()
    await restartMonitoring()
  }
}, { deep: true })

// Watch for custom path toggle changes
watch(() => config.value.useCustomPath, async (newValue) => {
  if (!newValue) {
    // User disabled custom path - clear the monitor path
    config.value.monitorPath = ''
  }
  await saveConfig()
}, { deep: true })

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
  // Only auto-save and restart monitoring if the path actually changed and is not empty
  if (newPath && newPath !== oldPath) {
    await saveConfig()
    await restartMonitoring()
  }
}, { deep: true })

watch(() => config.value.maxFileSize, debouncedSave, { deep: true })
watch(() => config.value.checkInterval, debouncedSave, { deep: true })
watch(() => config.value.startupInitialization, debouncedSave, { deep: true })

// Watch for recursive monitoring changes and immediately restart the watcher
watch(() => config.value.recursiveMonitoring, async () => {
  await saveConfig()
  // Restart the file watcher with the new recursive setting for all active paths
  await restartMonitoring()
}, { deep: true })

// Version and update functions
const loadVersionInfo = async () => {
  try {
    versionInfo.value = await GetVersionInfo()
  } catch (err) {
    console.error('Failed to load version info:', err)
  }
}

const checkForUpdates = async () => {
  isCheckingUpdates.value = true
  try {
    updateInfo.value = await CheckForUpdates()
  } catch (err) {
    console.error('Failed to check for updates:', err)
    updateInfo.value = {
      available: false,
      error: 'Failed to check for updates: ' + err.message
    }
  } finally {
    isCheckingUpdates.value = false
  }
}

const openUpdateUrl = async () => {
  if (updateInfo.value && updateInfo.value.releaseURL) {
    try {
      await OpenUpdateURL(updateInfo.value.releaseURL)
    } catch (err) {
      console.error('Failed to open update URL:', err)
    }
  }
}

const toggleVersionDetails = () => {
  showVersionDetails.value = !showVersionDetails.value
}

// Load version info on mount
onMounted(() => {
  // Reset animations flag first
  animateNow.value = false
  
  // Load config data and version info with a small delay
  setTimeout(() => {
    loadConfig()
    loadVersionInfo()
  }, 10)
})
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
                </h3>                <div class="form-group">
                  <label for="monitorPath">Monitor Path</label>
                    <!-- Path Selection Cards -->
                  <div class="path-selection-container">
                    <div class="path-checkbox-group">
                      <!-- MedalTV Card -->
                      <div class="path-checkbox-card" :class="{ selected: config.useMedalTVPath }">
                        <input
                          v-model="config.useMedalTVPath"
                          type="checkbox"
                          class="path-checkbox-input"
                          id="medalTVPath"
                        />
                        <div class="path-checkbox-header">
                          <div class="path-checkbox-icon">M</div>
                          <div class="path-checkbox-title">MedalTV</div>
                        </div>
                        <div class="path-checkbox-description">
                          Automatically monitor MedalTV's configured clip folder
                        </div>
                      </div>
                        <!-- NVIDIA Card -->
                      <div class="path-checkbox-card" :class="{ selected: config.useNVIDIAPath }">
                        <input
                          v-model="config.useNVIDIAPath"
                          type="checkbox"
                          class="path-checkbox-input"
                          id="nvidiaPath"
                        />
                        <div class="path-checkbox-header">
                          <div class="path-checkbox-icon nvidia-icon">N</div>
                          <div class="path-checkbox-title">NVIDIA</div>
                        </div>
                        <div class="path-checkbox-description">
                          Automatically monitor NVIDIA's current directory
                        </div>
                      </div>
                    </div>
                  </div>
                    <!-- Custom Path Toggle -->
                  <div class="form-group">
                    <label for="useCustomPath">Use custom folder path</label>
                    <div class="toggle-container">
                      <label class="toggle-switch">
                        <input
                          id="useCustomPath"
                          v-model="config.useCustomPath"
                          type="checkbox"
                          class="toggle-input"
                        />
                        <span class="toggle-slider"></span>
                      </label>
                      <span class="toggle-label">{{ config.useCustomPath ? 'Enabled' : 'Disabled' }}</span>
                    </div>
                    <p class="form-help">
                      Enable this to select an additional custom folder path for monitoring
                    </p>
                  </div>                  <!-- Custom Path Input -->
                  <div class="input-group" v-if="config.useCustomPath">
                    <input
                      id="monitorPath"
                      v-model="config.monitorPath"
                      type="text"
                      placeholder="Select a folder to monitor"
                      class="form-input"
                    />
                    <button 
                      @click="selectFolder" 
                      class="folder-button"
                    >
                      <FolderOpen :size="16" />
                      Browse
                    </button>
                  </div>                <p class="form-help">
                    <template v-if="config.useMedalTVPath && config.useNVIDIAPath && config.useCustomPath">
                      Monitoring MedalTV, NVIDIA, and custom paths
                    </template>
                    <template v-else-if="config.useMedalTVPath && config.useNVIDIAPath">
                      Monitoring both MedalTV and NVIDIA paths
                    </template>
                    <template v-else-if="config.useMedalTVPath && config.useCustomPath">
                      Monitoring MedalTV and custom paths
                    </template>
                    <template v-else-if="config.useNVIDIAPath && config.useCustomPath">
                      Monitoring NVIDIA and custom paths
                    </template>
                    <template v-else-if="config.useMedalTVPath">
                      Path automatically set from MedalTV settings
                    </template>
                    <template v-else-if="config.useNVIDIAPath">
                      Path automatically set from NVIDIA settings
                    </template>
                    <template v-else-if="config.useCustomPath">
                      Custom folder path selected for monitoring
                    </template>
                    <template v-else>
                      Select one or more monitoring options above
                    </template>
                  </p><!-- Paths Preview -->
                  <div class="paths-preview" v-if="(config.useMedalTVPath || config.useNVIDIAPath) || (config.useCustomPath && config.monitorPath)">
                    <div class="paths-preview-header">
                      <h4>Active Monitoring Paths:</h4>
                    </div>
                    <div class="paths-preview-list">
                      <div class="preview-path medaltu-preview" v-if="config.useMedalTVPath">
                        <div class="preview-path-icon medaltu-icon">M</div>
                        <div class="preview-path-info">
                          <div class="preview-path-label">MedalTV</div>
                          <div class="preview-path-value">{{ medalTVPath || 'Loading...' }}</div>
                        </div>
                      </div>
                      
                      <div class="preview-path nvidia-preview" v-if="config.useNVIDIAPath">
                        <div class="preview-path-icon nvidia-icon">N</div>
                        <div class="preview-path-info">
                          <div class="preview-path-label">NVIDIA</div>
                          <div class="preview-path-value">{{ nvidiaPath || 'Loading...' }}</div>
                        </div>
                      </div>
                        <div class="preview-path custom-preview" v-if="config.useCustomPath && config.monitorPath">
                        <div class="preview-path-icon custom-icon">C</div>
                        <div class="preview-path-info">
                          <div class="preview-path-label">Custom Path</div>
                          <div class="preview-path-value">{{ config.monitorPath }}</div>
                        </div>
                      </div>
                    </div>
                  </div>
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
                
                <div class="form-group">
                  <label for="desktopShortcut">Create desktop shortcut</label>
                  <div class="toggle-container">
                    <label class="toggle-switch">
                      <input
                        id="desktopShortcut"
                        v-model="config.desktopShortcut"
                        type="checkbox"
                        class="toggle-input"
                      />
                      <span class="toggle-slider"></span>
                    </label>
                    <span class="toggle-label">{{ config.desktopShortcut ? 'Enabled' : 'Disabled' }}</span>
                  </div>
                  <p class="form-help">
                    Create and maintain a desktop shortcut for AutoClipSend
                  </p>
                </div>
              </div>
            </transition>
            
            <!-- Version Information -->
            <transition name="fade-slide-up" :duration="{ enter: 800, leave: 600 }" appear>
              <div class="config-section info" v-if="animateNow">
                <h3>
                  <Info :size="18" />
                  Version Information
                </h3>
                
                <div class="version-display">
                  <div class="version-main">
                    <span class="version-number">{{ versionInfo.version || 'Loading...' }}</span>
                    <button @click="toggleVersionDetails" class="version-toggle">
                      <span v-if="!showVersionDetails">Show Details</span>
                      <span v-else>Hide Details</span>
                    </button>
                  </div>
                  
                  <transition name="fade-slide-down">
                    <div class="version-details" v-if="showVersionDetails">
                      <div class="version-item">
                        <strong>Build Date:</strong> {{ versionInfo.formattedDate || 'Unknown' }}
                      </div>
                      <div class="version-item">
                        <strong>Commit:</strong> {{ versionInfo.shortCommit || 'Unknown' }}
                      </div>
                      <div class="version-item">
                        <strong>Go Version:</strong> {{ versionInfo.goVersion || 'Unknown' }}
                      </div>
                    </div>
                  </transition>
                  
                  <div class="update-section">
                    <button @click="checkForUpdates" :disabled="isCheckingUpdates" class="btn-update">
                      <RefreshCw :size="16" :class="{ 'spinning': isCheckingUpdates }" />
                      {{ isCheckingUpdates ? 'Checking...' : 'Check for Updates' }}
                    </button>
                    
                    <div v-if="updateInfo" class="update-info">
                      <div v-if="updateInfo.available" class="update-available">
                        <div class="update-message">
                          <Download :size="16" />
                          <span>Update Available: {{ updateInfo.latestVersion }}</span>
                        </div>
                        <button @click="openUpdateUrl" class="btn-download">
                          <ExternalLink :size="16" />
                          Download Update
                        </button>
                      </div>
                      <div v-else-if="!updateInfo.error" class="update-current">
                        <span>✅ You're running the latest version</span>
                      </div>
                      <div v-if="updateInfo.error" class="update-error">
                        <span>⚠️ {{ updateInfo.error }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </transition>
          </div>
        </div>
      </div></transition>
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
  border-radius: 12px;
  padding: 1.5rem;
  border: 1px solid var(--border-default);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
  position: relative;
  transition: border-color 0.3s ease,
              box-shadow 0.3s ease;
}

.config-section:hover {
  border-color: var(--border-light);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.config-section.primary {
  border-color: var(--border-accent);
}

.config-section.secondary {
  border-color: rgba(138, 43, 226, 0.4);
}

.config-section.secondary:hover {
  border-color: rgba(138, 43, 226, 0.6);
}

.config-section h3 svg {
  color: var(--primary-color);
  transition: color 0.3s ease;
}

.config-section:hover h3 svg {
  color: var(--primary-light);
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
  transition: border-color 0.3s ease, box-shadow 0.3s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
}

.form-input:focus {
  outline: none;
  border-color: #ff8c00;
  box-shadow: 0 0 0 2px rgba(255, 140, 0, 0.2);
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
  transition: background-color 0.3s ease, transform 0.2s ease;
  white-space: nowrap;
  display: flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.8rem;
}

.test-button:hover, .update-button:hover, .folder-button:hover, .update-path-button:hover {
  background: #e67e22;
  transform: translateY(-1px);
}

.test-button:disabled, .update-button:disabled, .update-path-button:disabled {
  background: rgba(255, 140, 0, 0.3);
  cursor: not-allowed;
}

.folder-button {
  background: #2196f3;
}

.folder-button:hover {
  background: #1976d2;
  transform: translateY(-1px);
}

.update-path-button {
  width: 100%;
  margin-top: 0.5rem;
  justify-content: center;
  background: #4CAF50;
}

.update-path-button:hover {
  background: #45a049;
  transform: translateY(-1px);
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
  transition: all 0.3s ease;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
}

.toggle-slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 2px;
  bottom: 2px;
  background-color: #ffffff;
  transition: all 0.3s ease;
  border-radius: 50%;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}

.toggle-input:checked + .toggle-slider {
  background-color: #ff8c00;
  border-color: #ff8c00;
}

.toggle-input:checked + .toggle-slider:before {
  transform: translateX(26px);
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

/* Enhanced Checkbox Styling for MedalTV and NVIDIA */
.path-selection-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin-bottom: 1rem;
}

.path-checkbox-group {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  margin-bottom: 1rem;
}

.path-checkbox-card {
  background: linear-gradient(135deg, var(--bg-elements), var(--bg-interactive));
  border: 2px solid transparent;
  border-radius: 12px;
  padding: 1rem;
  transition: all 0.3s ease;
  cursor: pointer;
  position: relative;
}

.path-checkbox-card:hover {
  border-color: var(--primary-alpha);
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(255, 124, 61, 0.2);
}

.path-checkbox-card.selected {
  border-color: var(--primary-color);
  background: linear-gradient(135deg, rgba(255, 124, 61, 0.1), rgba(255, 124, 61, 0.05));
  box-shadow: 0 4px 16px rgba(255, 124, 61, 0.3);
}

.path-checkbox-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
}

.path-checkbox-icon {
  width: 24px;
  height: 24px;
  background: var(--primary-color);
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: bold;
  font-size: 0.8rem;
}

.path-checkbox-icon.nvidia-icon {
  background: #4ade80; /* Green for NVIDIA */
}

.path-checkbox-title {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 0.95rem;
}

.path-checkbox-description {
  color: var(--text-muted);
  font-size: 0.8rem;
  line-height: 1.4;
  margin-bottom: 0.75rem;
}

.path-checkbox-input {
  position: absolute;
  top: 1rem;
  right: 1rem;
  width: 18px;
  height: 18px;
  accent-color: var(--primary-color);
  cursor: pointer;
}

@media (max-width: 768px) {
  .path-checkbox-group {
    grid-template-columns: 1fr;
  }
}

/* Paths Preview Styles */
.paths-preview {
  margin-top: 1rem;
  padding: 1rem;
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid var(--border-default);
  border-radius: 8px;
}

.paths-preview-header {
  margin-bottom: 0.75rem;
}

.paths-preview-header h4 {
  color: var(--text-primary);
  font-size: 0.9rem;
  font-weight: 600;
  margin: 0;
}

.paths-preview-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.preview-path {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 0.75rem;
  background: var(--bg-elements);
  border-radius: 6px;
  border: 1px solid var(--border-default);
  transition: var(--transition-smooth);
}

.preview-path:hover {
  border-color: var(--border-light);
  background: var(--bg-interactive);
}

.preview-path-icon {
  width: 20px;
  height: 20px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: bold;
  font-size: 0.75rem;
  flex-shrink: 0;
  margin-top: 0.1rem;
}

.preview-path-icon.medaltu-icon {
  background: var(--primary-color);
}

.preview-path-icon.nvidia-icon {
  background: #4ade80;
}

.preview-path-info {
  flex: 1;
  min-width: 0;
}

.preview-path-label {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 0.2rem;
}

.preview-path-value {
  font-size: 0.75rem;
  color: var(--text-muted);
  font-family: 'Courier New', monospace;
  word-break: break-all;
  line-height: 1.3;
}

.medaltu-preview {
  border-color: rgba(255, 124, 61, 0.3);
}

.nvidia-preview {
  border-color: rgba(74, 222, 128, 0.3);
}

.custom-preview {
  border-color: rgba(59, 130, 246, 0.3);
}

.preview-path-icon.custom-icon {
  background: #3b82f6; /* Blue for custom */
}

/* Version Information Styles */
.version-display {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.version-main {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem;
  background: rgba(0, 0, 0, 0.2);
  border-radius: 8px;
  border: 1px solid var(--border-default);
}

.version-number {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--primary-color);
  font-family: 'Courier New', monospace;
}

.version-toggle {
  background: transparent;
  border: 1px solid var(--border-light);
  color: var(--text-muted);
  padding: 0.4rem 0.75rem;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.8rem;
  transition: var(--transition-smooth);
}

.version-toggle:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.version-details {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 0.75rem;
  background: rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  border: 1px solid var(--border-default);
}

.version-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.version-item strong {
  color: var(--text-primary);
  font-weight: 600;
}

.update-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.btn-update {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  background: linear-gradient(135deg, var(--primary-color), var(--primary-dark));
  border: none;
  border-radius: 8px;
  color: white;
  font-weight: 600;
  cursor: pointer;
  transition: var(--transition-smooth);
  font-size: 0.9rem;
}

.btn-update:hover {
  background: linear-gradient(135deg, var(--primary-light), var(--primary-color));
  transform: translateY(-1px);
  box-shadow: var(--shadow-medium);
}

.btn-update:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.update-info {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.update-available {
  background: linear-gradient(135deg, rgba(74, 222, 128, 0.1), rgba(74, 222, 128, 0.05));
  border: 1px solid rgba(74, 222, 128, 0.3);
  border-radius: 8px;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.update-message {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #4ade80;
  font-weight: 600;
  font-size: 0.9rem;
}

.btn-download {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.6rem 1rem;
  background: linear-gradient(135deg, #4ade80, #22c55e);
  border: none;
  border-radius: 6px;
  color: white;
  font-weight: 600;
  cursor: pointer;
  transition: var(--transition-smooth);
  font-size: 0.85rem;
}

.btn-download:hover {
  background: linear-gradient(135deg, #22c55e, #16a34a);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(74, 222, 128, 0.3);
}

.update-current {
  color: #4ade80;
  font-size: 0.9rem;
  font-weight: 500;
  padding: 0.5rem;
  text-align: center;
}

.update-error {
  color: var(--error-color);
  font-size: 0.85rem;
  padding: 0.5rem;
  text-align: center;
}

.fade-slide-down-enter-active,
.fade-slide-down-leave-active {
  transition: all 0.3s ease;
}

.fade-slide-down-enter-from,
.fade-slide-down-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
