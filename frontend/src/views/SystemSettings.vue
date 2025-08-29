<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import apiClient from '@/api/client';

const systemInfo = ref({
  firstTrafficFile: null,
  lastTrafficFile: null,
  trafficTimeRange: null,
  lastUpdate: null
});

const loading = ref(false);
let refreshInterval = null;

async function fetchSystemInfo() {
  loading.value = true;
  try {
    const response = await apiClient.get('/system/info');
    systemInfo.value = response.data;
    systemInfo.value.lastUpdate = new Date();
  } catch (error) {
    console.error('Ошибка получения информации о системе:', error);
  } finally {
    loading.value = false;
  }
}

function formatDate(dateString) {
  if (!dateString) return 'Неизвестно';
  return new Date(dateString).toLocaleString('ru-RU');
}

function formatTimeRange() {
  const info = systemInfo.value;
  if (!info.firstTrafficFile || !info.lastTrafficFile) return 'Неизвестно';
  
  const first = new Date(info.firstTrafficFile);
  const last = new Date(info.lastTrafficFile);
  const diffMs = last - first;
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
  const diffHours = Math.floor((diffMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
  
  return `${diffDays} дней ${diffHours} часов`;
}

onMounted(() => {
  fetchSystemInfo();
  // Автоматическое обновление каждые 5 секунд
  refreshInterval = setInterval(fetchSystemInfo, 5000);
});

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval);
  }
});
</script>

<template>
  <div class="page-container">
    <header class="page-header">
      <h1>Настройки системы</h1>
    </header>

    <div class="settings-content">
      <div class="card">
        <div class="card-header">
          <h2>Информация о трафике</h2>
          <div class="last-update">
            Последнее обновление: {{ systemInfo.lastUpdate ? formatDate(systemInfo.lastUpdate) : 'Неизвестно' }}
          </div>
        </div>
        
        <div v-if="loading" class="loading-container">
          <div class="spinner"></div>
          <span>Загрузка информации о системе...</span>
        </div>
        
        <div v-else class="info-grid">
          <div class="info-item">
            <label>Первый файл трафика:</label>
            <span class="value">{{ formatDate(systemInfo.firstTrafficFile) }}</span>
          </div>
          
          <div class="info-item">
            <label>Последний файл трафика:</label>
            <span class="value">{{ formatDate(systemInfo.lastTrafficFile) }}</span>
          </div>
          
          <div class="info-item">
            <label>Доступный период трафика:</label>
            <span class="value">{{ formatTimeRange() }}</span>
          </div>
        </div>
      </div>
      
      <div class="card">
        <div class="card-header">
          <h2>Системная информация</h2>
        </div>
        
        <div class="info-grid">
          <div class="info-item">
            <label>Версия системы:</label>
            <span class="value">1.0.0</span>
          </div>
          
          <div class="info-item">
            <label>Время работы:</label>
            <span class="value">Доступно в следующих версиях</span>
          </div>
          
          <div class="info-item">
            <label>Автообновление:</label>
            <span class="value auto-refresh">
              <span class="material-icons icon-sm">sync</span>
              Каждые 5 секунд
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.settings-content {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--gray-200);
}

.card-header h2 {
  margin: 0;
  color: var(--gray-900);
  font-size: 1.25rem;
  font-weight: 600;
}

.last-update {
  color: var(--gray-600);
  font-size: 0.875rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  padding: 3rem;
}

.info-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.5rem;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background: var(--gray-50);
  border-radius: var(--radius-lg);
  border-left: 4px solid var(--primary-500);
}

.info-item label {
  font-weight: 500;
  color: var(--gray-700);
}

.info-item .value {
  font-weight: 600;
  color: var(--gray-900);
}

.auto-refresh {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--success-600);
}

.auto-refresh .material-icons {
  animation: spin 2s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@media (min-width: 768px) {
  .info-grid {
    grid-template-columns: 1fr 1fr;
  }
}

@media (min-width: 1024px) {
  .settings-content {
    max-width: 1200px;
  }
}
</style>