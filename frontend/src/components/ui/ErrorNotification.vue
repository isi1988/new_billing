<script setup>
import { ref, watch } from 'vue';
import { useNotificationStore } from '@/stores/notification';

const notificationStore = useNotificationStore();

const hideNotification = (id) => {
  notificationStore.removeNotification(id);
};

const getNotificationIcon = (type) => {
  switch (type) {
    case 'error': return 'error';
    case 'warning': return 'warning';
    case 'success': return 'check_circle';
    case 'info': return 'info';
    default: return 'info';
  }
};

const getNotificationClass = (type) => {
  switch (type) {
    case 'error': return 'notification-error';
    case 'warning': return 'notification-warning';
    case 'success': return 'notification-success';
    case 'info': return 'notification-info';
    default: return 'notification-info';
  }
};
</script>

<template>
  <div class="notifications-container">
    <div
      v-for="notification in notificationStore.notifications"
      :key="notification.id"
      :class="['notification', getNotificationClass(notification.type)]"
    >
      <div class="notification-content">
        <div class="notification-header">
          <span class="material-icons icon-sm">{{ getNotificationIcon(notification.type) }}</span>
          <h4 class="notification-title">{{ notification.title }}</h4>
          <button 
            @click="hideNotification(notification.id)"
            class="notification-close"
          >
            <span class="material-icons icon-sm">close</span>
          </button>
        </div>
        <div v-if="notification.message" class="notification-message">
          {{ notification.message }}
        </div>
        <div v-if="notification.details" class="notification-details">
          <details>
            <summary>Подробности</summary>
            <pre>{{ notification.details }}</pre>
          </details>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.notifications-container {
  position: fixed;
  top: 1rem;
  right: 1rem;
  z-index: 10000;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-width: 400px;
}

.notification {
  background: white;
  border-radius: var(--radius-lg);
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.15);
  overflow: hidden;
  animation: slideIn 0.3s ease-out;
  border-left: 4px solid;
}

@keyframes slideIn {
  from {
    transform: translateX(100%);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

.notification-error {
  border-left-color: var(--error-500);
}

.notification-warning {
  border-left-color: var(--warning-500);
}

.notification-success {
  border-left-color: var(--success-500);
}

.notification-info {
  border-left-color: var(--info-500);
}

.notification-content {
  padding: 1rem;
}

.notification-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.notification-title {
  flex: 1;
  margin: 0;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--gray-900);
}

.notification-close {
  background: none;
  border: none;
  color: var(--gray-400);
  cursor: pointer;
  padding: 0.25rem;
  border-radius: var(--radius-md);
  transition: all 0.2s ease;
}

.notification-close:hover {
  background: var(--gray-100);
  color: var(--gray-600);
}

.notification-message {
  font-size: 0.875rem;
  color: var(--gray-700);
  line-height: 1.4;
  margin-bottom: 0.5rem;
}

.notification-details {
  margin-top: 0.5rem;
}

.notification-details summary {
  font-size: 0.75rem;
  color: var(--gray-600);
  cursor: pointer;
  padding: 0.25rem 0;
}

.notification-details pre {
  font-size: 0.75rem;
  color: var(--gray-600);
  background: var(--gray-50);
  padding: 0.5rem;
  border-radius: var(--radius-md);
  white-space: pre-wrap;
  word-break: break-all;
  margin-top: 0.5rem;
  max-height: 200px;
  overflow-y: auto;
}

.notification-error .material-icons {
  color: var(--error-500);
}

.notification-warning .material-icons {
  color: var(--warning-500);
}

.notification-success .material-icons {
  color: var(--success-500);
}

.notification-info .material-icons {
  color: var(--info-500);
}
</style>