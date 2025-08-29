import { reactive } from 'vue';

// Global reactive store
const state = reactive({
  notifications: []
});

let nextId = 1;

const addNotification = (notification) => {
  const id = nextId++;
  const newNotification = {
    id,
    type: notification.type || 'info',
    title: notification.title || 'Уведомление',
    message: notification.message,
    details: notification.details,
    duration: notification.duration || (notification.type === 'error' ? 10000 : 5000)
  };

  state.notifications.push(newNotification);

  // Auto remove after duration
  if (newNotification.duration > 0) {
    setTimeout(() => {
      removeNotification(id);
    }, newNotification.duration);
  }

  return id;
};

const removeNotification = (id) => {
  const index = state.notifications.findIndex(n => n.id === id);
  if (index > -1) {
    state.notifications.splice(index, 1);
  }
};

const clearAll = () => {
  state.notifications = [];
};

// Helper functions for different types
const showError = (title, message, details) => {
  return addNotification({
    type: 'error',
    title,
    message,
    details,
    duration: 10000
  });
};

const showSuccess = (title, message) => {
  return addNotification({
    type: 'success',
    title,
    message,
    duration: 3000
  });
};

const showWarning = (title, message) => {
  return addNotification({
    type: 'warning',
    title,
    message,
    duration: 5000
  });
};

const showInfo = (title, message) => {
  return addNotification({
    type: 'info',
    title,
    message,
    duration: 5000
  });
};

export const useNotificationStore = () => {
  return {
    notifications: state.notifications,
    addNotification,
    removeNotification,
    clearAll,
    showError,
    showSuccess,
    showWarning,
    showInfo
  };
};