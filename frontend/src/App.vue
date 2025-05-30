<script setup>
import { ref, onMounted } from 'vue'
import { Activity, Settings } from 'lucide-vue-next'
import StatusPage from './components/StatusPage.vue'
import ConfigPage from './components/ConfigPage.vue'
import Notification from './components/Notification.vue'
import { EventsOn } from '../wailsjs/runtime/runtime'

const currentPage = ref('status')

const switchPage = (page) => {
  currentPage.value = page
}

onMounted(() => {
  // Only listen for app state changes here.
  EventsOn('app-minimized-to-tray', () => {
    console.log('App minimized to tray')
  })
  EventsOn('app-restored-from-tray', () => {
    console.log('App restored from tray')
  })
})
</script>

<template>
  <div id="app">
    <nav class="navigation">
      <div class="nav-brand">
        <h1>AutoClipSend</h1>
      </div>
      <div class="nav-menu">
        <button 
          :class="{ active: currentPage === 'status' }"
          @click="switchPage('status')"
          class="nav-button"
        >
          <Activity :size="16" />
          Status
        </button>
        <button 
          :class="{ active: currentPage === 'config' }"
          @click="switchPage('config')"
          class="nav-button"
        >
          <Settings :size="16" />
          Settings
        </button>
      </div>
    </nav>

    <main class="main-content">
      <StatusPage v-if="currentPage === 'status'" />
      <ConfigPage v-if="currentPage === 'config'" />
    </main>
    
    <!-- Notification Modal - Always present to catch events -->
    <Notification />
  </div>
</template>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
  background: #1a1a1a;
  color: #ffffff;
  overflow: hidden;
}

#app {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: linear-gradient(135deg, #1a1a1a 0%, #2d2d2d 100%);
}

.navigation {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1.5rem;
  background: rgba(0, 0, 0, 0.3);
  border-bottom: 1px solid rgba(255, 140, 0, 0.3);
  backdrop-filter: blur(10px);
  flex-shrink: 0;
}

.nav-brand h1 {
  color: #ff8c00;
  font-size: 1.3rem;
  font-weight: 600;
}

.nav-menu {
  display: flex;
  gap: 0.75rem;
}

.nav-button {
  padding: 0.5rem 1rem;
  background: transparent;
  border: 1px solid rgba(255, 140, 0, 0.3);
  color: #ffffff;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.3s ease;
  font-size: 0.85rem;
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.nav-button:hover {
  background: rgba(255, 140, 0, 0.1);
  border-color: rgba(255, 140, 0, 0.6);
}

.nav-button.active {
  background: #ff8c00;
  color: #1a1a1a;
  border-color: #ff8c00;
}

.main-content {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
