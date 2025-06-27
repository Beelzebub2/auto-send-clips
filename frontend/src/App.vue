<script setup>
import { ref, onMounted } from 'vue'
import { Activity, Settings, Film } from 'lucide-vue-next'
import StatusPage from './components/StatusPage.vue'
import ConfigPage from './components/ConfigPage.vue'
import ClipsPage from './components/ClipsPage.vue'
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
          :class="{ active: currentPage === 'clips' }"
          @click="switchPage('clips')"
          class="nav-button"
        >
          <Film :size="16" />
          Clips
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
      <ClipsPage v-if="currentPage === 'clips'" />
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
  background: linear-gradient(135deg, var(--bg-darkest) 0%, var(--bg-sections) 100%);
  overflow: hidden;
}

.navigation {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.6rem 1.2rem;
  background: var(--bg-sections);
  border-bottom: 1px solid var(--border-default);
  backdrop-filter: blur(20px);
  flex-shrink: 0;
  box-shadow: var(--shadow-small);
  min-height: 60px;
  max-height: 60px;
}

.nav-brand {
  flex-shrink: 0;
}

.nav-brand h1 {
  color: var(--primary-color);
  font-size: 1.3rem;
  font-weight: 700;
  margin: 0;
  margin-right: 1.5rem;
  text-shadow: 0 2px 8px rgba(255, 124, 61, 0.3);
  white-space: nowrap;
}

.nav-menu {
  display: flex;
  gap: 0.6rem;
  flex-shrink: 0;
}

.nav-button {
  padding: 0.5rem 1rem;
  background: var(--bg-elements);
  border: 1px solid var(--border-default);
  color: var(--text-secondary);
  border-radius: 8px;
  cursor: pointer;
  transition: var(--transition-smooth);
  font-size: 0.8rem;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 0.4rem;
  box-shadow: var(--shadow-small);
  white-space: nowrap;
  min-width: fit-content;
}

.nav-button:hover {
  background: var(--bg-interactive);
  border-color: var(--border-accent);
  transform: translateY(-2px);
  box-shadow: var(--shadow-medium);
  color: var(--text-primary);
}

.nav-button.active {
  background: linear-gradient(135deg, var(--primary-color), var(--primary-dark));
  color: #ffffff;
  border-color: var(--primary-color);
  box-shadow: 0 4px 16px rgba(255, 124, 61, 0.4);
  font-weight: 600;
}

.nav-button.active:hover {
  transform: translateY(-2px) scale(1.02);
  box-shadow: 0 6px 20px rgba(255, 124, 61, 0.5);
}

.main-content {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: var(--bg-darkest);
}

/* Responsive Navigation */
@media (max-width: 600px) {
  .navigation {
    padding: 0.5rem 0.8rem;
  }
  
  .nav-brand h1 {
    font-size: 1.1rem;
    margin-right: 1rem;
  }
  
  .nav-menu {
    gap: 0.4rem;
  }
  
  .nav-button {
    padding: 0.4rem 0.8rem;
    font-size: 0.75rem;
    gap: 0.3rem;
  }
  
  .nav-button svg {
    width: 14px;
    height: 14px;
  }
}

@media (max-width: 400px) {
  .navigation {
    padding: 0.4rem 0.6rem;
  }
  
  .nav-brand h1 {
    font-size: 1rem;
    margin-right: 0.5rem;
  }
  
  .nav-button {
    padding: 0.4rem 0.6rem;
    font-size: 0.7rem;
  }
  
  .nav-button span {
    display: none;
  }
}
</style>
