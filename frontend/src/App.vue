<script setup>
import { ref } from 'vue';
import AppHeader from '@/components/AppHeader.vue';
import SideDrawer from '@/components/SideDrawer.vue';

const isDrawerOpen = ref(false);
</script>

<template>
  <div class="app-layout">
    <AppHeader @toggle-drawer="isDrawerOpen = !isDrawerOpen" />
    <SideDrawer :is-open="isDrawerOpen" @close="isDrawerOpen = false" />

    <main class="main-content">
      <!-- Здесь будут отображаться наши страницы: UserManagement, TariffManagement и т.д. -->
      <router-view />
    </main>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  flex-direction: column;
  height: 100vh; /* Занимает всю высоту экрана */
  overflow: hidden; /* Предотвращает двойной скроллбар */
}

.main-content {
  flex-grow: 1; /* Занимает все оставшееся пространство */
  overflow-y: auto; /* Позволяет прокручивать только основное содержимое */
  background-color: var(--background-color);
}
</style>