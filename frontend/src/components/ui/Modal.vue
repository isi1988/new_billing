<script setup>
defineProps({
  isOpen: Boolean,
});
defineEmits(['close']);
</script>

<template>
  <teleport to="body">
    <div v-if="isOpen" class="modal-overlay" @click="$emit('close')">
      <div class="modal" @click.stop>
        <header class="modal-header">
          <div class="modal-title">
            <slot name="header"><h2>Заголовок</h2></slot>
          </div>
          <button @click="$emit('close')" class="modal-close">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
            </svg>
          </button>
        </header>
        <main class="modal-body">
          <slot></slot>
        </main>
        <footer v-if="$slots.footer" class="modal-footer">
          <slot name="footer"></slot>
        </footer>
      </div>
    </div>
  </teleport>
</template>

<style scoped>
.w-5 {
  width: 1.25rem;
  height: 1.25rem;
}
</style>