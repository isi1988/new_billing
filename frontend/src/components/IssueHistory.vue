<script setup>
import { ref, onMounted, watch } from 'vue';
import apiClient from '@/api/client';

const props = defineProps({
  issueId: { type: Number, required: true },
  isOpen: { type: Boolean, default: false },
});

const emit = defineEmits(['close']);

const history = ref([]);
const loading = ref(false);

async function fetchHistory() {
  if (!props.issueId) return;
  
  loading.value = true;
  try {
    const response = await apiClient.get(`/issues/${props.issueId}/history`);
    history.value = response.data || [];
  } catch (error) {
    console.error('Failed to fetch issue history:', error);
    history.value = [];
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  if (props.isOpen) {
    fetchHistory();
  }
});

// Watch for changes in isOpen prop
watch(() => props.isOpen, (newValue) => {
  if (newValue) {
    fetchHistory();
  }
});

function formatFieldName(fieldName) {
  switch (fieldName) {
    case 'title': return 'Название';
    case 'description': return 'Описание';
    default: return fieldName;
  }
}

function formatDate(dateString) {
  return new Date(dateString).toLocaleString('ru-RU');
}
</script>

<template>
  <div v-if="isOpen" class="history-overlay" @click="$emit('close')">
    <div class="history-modal" @click.stop>
      <div class="history-header">
        <h3>История изменений</h3>
        <button class="close-button" @click="$emit('close')">
          <svg class="close-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>

      <div class="history-content">
        <div v-if="loading" class="loading">
          Загружаем историю изменений...
        </div>

        <div v-else-if="history.length === 0" class="empty-state">
          История изменений пуста
        </div>

        <div v-else class="history-list">
          <div 
            v-for="change in history" 
            :key="change.id" 
            class="history-item"
          >
            <div class="history-meta">
              <span class="history-field">{{ formatFieldName(change.field_name) }}</span>
              <span class="history-date">{{ formatDate(change.edited_at) }}</span>
            </div>
            
            <div class="history-changes">
              <div class="change-section">
                <label>Было:</label>
                <div class="old-value">{{ change.old_value || '(пусто)' }}</div>
              </div>
              
              <div class="change-arrow">→</div>
              
              <div class="change-section">
                <label>Стало:</label>
                <div class="new-value">{{ change.new_value || '(пусто)' }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.history-overlay {
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
}

.history-modal {
  background: white;
  border-radius: 0.5rem;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  max-width: 800px;
  width: 90%;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid var(--gray-200);
}

.history-header h3 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--gray-900);
}

.close-button {
  padding: 0.5rem;
  background: none;
  border: none;
  color: var(--gray-500);
  cursor: pointer;
  border-radius: 0.25rem;
  transition: all 0.2s ease;
}

.close-button:hover {
  background-color: var(--gray-100);
  color: var(--gray-700);
}

.close-icon {
  width: 1.25rem;
  height: 1.25rem;
}

.history-content {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem;
}

.loading, .empty-state {
  text-align: center;
  padding: 2rem;
  color: var(--gray-500);
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.history-item {
  border: 1px solid var(--gray-200);
  border-radius: 0.5rem;
  padding: 1rem;
  background: var(--gray-50);
}

.history-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--gray-300);
}

.history-field {
  font-weight: 600;
  color: var(--primary-700);
  font-size: 0.875rem;
}

.history-date {
  font-size: 0.75rem;
  color: var(--gray-500);
}

.history-changes {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  gap: 1rem;
  align-items: start;
}

.change-section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.change-section label {
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--gray-600);
  text-transform: uppercase;
}

.old-value, .new-value {
  padding: 0.75rem;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
}

.old-value {
  background-color: #fef2f2;
  color: #991b1b;
  border: 1px solid #fecaca;
}

.new-value {
  background-color: #f0fdf4;
  color: #166534;
  border: 1px solid #bbf7d0;
}

.change-arrow {
  align-self: center;
  color: var(--gray-400);
  font-size: 1.25rem;
  font-weight: bold;
}

@media (max-width: 768px) {
  .history-modal {
    width: 95%;
    max-height: 90vh;
  }

  .history-changes {
    grid-template-columns: 1fr;
    gap: 0.75rem;
  }

  .change-arrow {
    justify-self: center;
    transform: rotate(90deg);
  }
}
</style>