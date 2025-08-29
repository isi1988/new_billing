<template>
  <div v-if="isOpen" class="confirm-dialog-overlay" @click="handleOverlayClick">
    <div class="confirm-dialog">
      <div class="confirm-dialog-header">
        <div class="confirm-icon" :class="`confirm-icon-${type}`">
          <span class="material-icons">{{ iconName }}</span>
        </div>
        <h3 class="confirm-title">{{ title }}</h3>
      </div>
      
      <div class="confirm-dialog-body">
        <p class="confirm-message">{{ message }}</p>
        <div v-if="details" class="confirm-details">
          {{ details }}
        </div>
      </div>
      
      <div class="confirm-dialog-actions">
        <button 
          @click="$emit('cancel')" 
          class="btn btn-secondary"
          type="button"
        >
          {{ cancelText }}
        </button>
        <button 
          @click="$emit('confirm')" 
          class="btn"
          :class="`btn-${buttonType}`"
          type="button"
        >
          {{ confirmText }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  isOpen: { type: Boolean, default: false },
  type: { type: String, default: 'warning' }, // 'warning', 'danger', 'info'
  title: { type: String, required: true },
  message: { type: String, required: true },
  details: { type: String, default: '' },
  confirmText: { type: String, default: 'Подтвердить' },
  cancelText: { type: String, default: 'Отмена' }
});

const iconName = computed(() => {
  switch (props.type) {
    case 'danger': return 'delete';
    case 'warning': return 'warning';
    case 'info': return 'info';
    default: return 'help';
  }
});

const buttonType = computed(() => {
  switch (props.type) {
    case 'danger': return 'danger';
    case 'warning': return 'warning';
    case 'info': return 'primary';
    default: return 'primary';
  }
});

const emit = defineEmits(['confirm', 'cancel']);

const handleOverlayClick = (event) => {
  if (event.target === event.currentTarget) {
    emit('cancel');
  }
};
</script>

<style scoped>
.confirm-dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.2s ease-out;
}

.confirm-dialog {
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  max-width: 400px;
  width: 90%;
  margin: 0 16px;
  animation: slideIn 0.2s ease-out;
}

.confirm-dialog-header {
  padding: 24px 24px 16px;
  text-align: center;
}

.confirm-icon {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
  position: relative;
}

.confirm-icon span {
  font-size: 28px;
  color: white;
}

.confirm-icon-danger {
  background: linear-gradient(135deg, var(--error-500) 0%, var(--error-600) 100%);
}

.confirm-icon-warning {
  background: linear-gradient(135deg, var(--warning-500) 0%, var(--warning-600) 100%);
}

.confirm-icon-info {
  background: linear-gradient(135deg, var(--primary-500) 0%, var(--primary-600) 100%);
}

.confirm-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--gray-900);
  margin: 0;
}

.confirm-dialog-body {
  padding: 0 24px 24px;
  text-align: center;
}

.confirm-message {
  font-size: 16px;
  color: var(--gray-700);
  line-height: 1.5;
  margin: 0 0 12px;
}

.confirm-details {
  font-size: 14px;
  color: var(--gray-600);
  background: var(--gray-50);
  border-radius: 6px;
  padding: 12px;
  border-left: 3px solid var(--gray-300);
}

.confirm-dialog-actions {
  padding: 0 24px 24px;
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.confirm-dialog-actions .btn {
  flex: 1;
  justify-content: center;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes slideIn {
  from { 
    opacity: 0;
    transform: translateY(-20px) scale(0.95);
  }
  to { 
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}
</style>