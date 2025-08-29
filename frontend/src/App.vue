<script setup>
import { ref } from 'vue';
import AppHeader from '@/components/AppHeader.vue';
import SideDrawer from '@/components/SideDrawer.vue';
import ErrorNotification from '@/components/ui/ErrorNotification.vue';

const isDrawerOpen = ref(false);
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <AppHeader @toggle-drawer="isDrawerOpen = !isDrawerOpen" />
    <SideDrawer :is-open="isDrawerOpen" @close="isDrawerOpen = false" />

    <main class="main-content">
      <div class="max-w-7xl mx-auto px-6 py-8">
        <router-view />
      </div>
    </main>

    <!-- Global error notifications -->
    <ErrorNotification />
  </div>
</template>

<style scoped>
.main-content {
  padding-top: 5rem; /* Account for fixed header height */
  min-height: calc(100vh - 5rem);
}
</style>