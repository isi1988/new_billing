<script setup>
defineProps({
  isOpen: Boolean,
});
defineEmits(['close']);
</script>

<template>
  <teleport to="body">
    <div v-if="isOpen" class="modal-overlay" @click="$emit('close')">
      <div class="modal-content" @click.stop>
        <header class="modal-header">
          <slot name="header"><h2>Заголовок</h2></slot>
          <button @click="$emit('close')" class="close-button">&times;</button>
        </header>
        <main class="modal-body">
          <slot></slot>
        </main>
      </div>
    </div>
  </teleport>
</template>

<style>
/* CSS для этого компонента уже есть в assets/css/components.css */
</style>