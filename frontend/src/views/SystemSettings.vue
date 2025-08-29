<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import apiClient from '@/api/client';
import DataTable from '@/components/ui/DataTable.vue';

const processedFiles = ref([]);
const pagination = ref({
  page: 1,
  limit: 25,
  total: 0,
  totalPages: 0
});
const loading = ref(false);
const systemInfo = ref({
  lastUpdate: null
});
let refreshInterval = null;

// Колонки для таблицы файлов
const columns = [
  { key: 'id', label: 'ID' },
  { key: 'file_name', label: 'Имя файла' },
  { 
    key: 'processed_at', 
    label: 'Дата обработки',
    formatter: (file) => formatDate(file.processed_at)
  }
];

async function fetchProcessedFiles() {
  loading.value = true;
  try {
    const response = await apiClient.get('/system/processed-files', {
      params: {
        page: pagination.value.page,
        limit: pagination.value.limit
      }
    });
    processedFiles.value = response.data.files || [];
    pagination.value = { ...pagination.value, ...response.data.pagination };
    systemInfo.value.lastUpdate = new Date();
  } catch (error) {
    console.error('Ошибка получения обработанных файлов:', error);
  } finally {
    loading.value = false;
  }
}

function formatDate(dateString) {
  if (!dateString) return 'Неизвестно';
  return new Date(dateString).toLocaleString('ru-RU');
}

function nextPage() {
  if (pagination.value.page < pagination.value.totalPages) {
    pagination.value.page++;
    fetchProcessedFiles();
  }
}

function previousPage() {
  if (pagination.value.page > 1) {
    pagination.value.page--;
    fetchProcessedFiles();
  }
}

function changePage(newPage) {
  pagination.value.page = newPage;
  fetchProcessedFiles();
}

onMounted(() => {
  fetchProcessedFiles();
  // Автоматическое обновление каждые 30 секунд
  refreshInterval = setInterval(fetchProcessedFiles, 30000);
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
          <h2>Обработанные файлы трафика</h2>
          <div class="last-update">
            <span class="material-icons icon-sm">sync</span>
            Последнее обновление: {{ systemInfo.lastUpdate ? formatDate(systemInfo.lastUpdate) : 'Неизвестно' }}
          </div>
        </div>
        
        <div v-if="loading" class="loading-container">
          <div class="spinner"></div>
          <span>Загрузка обработанных файлов...</span>
        </div>
        
        <div v-else>
          <DataTable
            :items="processedFiles"
            :columns="columns"
            :loading="loading"
            :show-actions="false"
          />
          
          <!-- Пагинация -->
          <div class="pagination-container">
            <div class="pagination-info">
              <span>Всего файлов: {{ pagination.total }}</span>
              <span>Страница {{ pagination.page }} из {{ pagination.totalPages }}</span>
            </div>
            <div class="pagination-controls">
              <button 
                @click="previousPage"
                :disabled="pagination.page <= 1"
                class="btn btn-sm btn-secondary"
              >
                <span class="material-icons icon-xs">arrow_back</span>
                Назад
              </button>
              <span class="page-numbers">
                <button 
                  v-for="page in Math.min(pagination.totalPages, 5)" 
                  :key="page"
                  @click="changePage(page)"
                  :class="['btn btn-sm', page === pagination.page ? 'btn-primary' : 'btn-secondary']"
                >
                  {{ page }}
                </button>
              </span>
              <button 
                @click="nextPage"
                :disabled="pagination.page >= pagination.totalPages"
                class="btn btn-sm btn-secondary"
              >
                Далее
                <span class="material-icons icon-xs">arrow_forward</span>
              </button>
            </div>
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
            <label>Автообновление файлов:</label>
            <span class="value auto-refresh">
              <span class="material-icons icon-sm">sync</span>
              Каждые 30 секунд
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

.pagination-container {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid var(--gray-200);
}

.pagination-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  color: var(--gray-600);
  font-size: 0.875rem;
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.page-numbers {
  display: flex;
  gap: 0.25rem;
  margin: 0 0.5rem;
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

@media (max-width: 768px) {
  .pagination-container {
    flex-direction: column;
    gap: 1rem;
    align-items: stretch;
  }
  
  .pagination-info {
    text-align: center;
  }
  
  .pagination-controls {
    justify-content: center;
  }
}

@media (min-width: 768px) {
  .info-grid {
    grid-template-columns: 1fr 1fr;
  }
  
  .pagination-info {
    flex-direction: row;
    gap: 2rem;
  }
}

@media (min-width: 1024px) {
  .settings-content {
    max-width: 1200px;
  }
}
</style>