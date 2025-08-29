<template>
  <div class="connection-stats">
    <div class="mb-6">
      <button 
        @click="$router.back()"
        class="btn btn-ghost mb-4"
      >
        ← Назад
      </button>
      <h1 class="dashboard-title">Статистика по подключению</h1>
    </div>

    <!-- Фильтры -->
    <div class="card mb-6">
      <div class="card-body">
        <h2 class="card-title">Период анализа</h2>
        <div class="filter-grid">
          <div class="form-group">
            <label class="form-label">Дата с</label>
            <input 
              v-model="filters.fromDate"
              type="date"
              class="form-control"
            />
          </div>
          <div class="form-group">
            <label class="form-label">Дата по</label>
            <input 
              v-model="filters.toDate"
              type="date"
              class="form-control"
            />
          </div>
        </div>
        <div class="mt-4">
          <button 
            @click="loadConnectionStats"
            class="btn btn-primary"
          >
            Обновить статистику
          </button>
        </div>
      </div>
    </div>

    <!-- Загрузка -->
    <div v-if="loading" class="loading-container">
      <div class="spinner"></div>
    </div>

    <!-- Ошибка -->
    <div v-else-if="error" class="alert alert-danger mb-6">
      <p>{{ error }}</p>
    </div>

    <!-- Основная статистика -->
    <div v-else-if="stats">
      <!-- Информация о подключении -->
      <div class="card mb-6">
        <div class="card-body">
          <h2 class="card-title">Информация о подключении</h2>
          <div class="connection-info-grid">
            <div class="connection-info-section">
              <p><strong>IP адрес:</strong> {{ stats.connection.ip_address }}/{{ stats.connection.mask }}</p>
              <p><strong>Адрес:</strong> {{ stats.connection.address }}</p>
              <p><strong>Тип подключения:</strong> {{ stats.connection.connection_type }}</p>
              <p><strong>Статус:</strong> 
                <span class="badge" :class="stats.connection.is_blocked ? 'badge-danger' : 'badge-success'">
                  {{ stats.connection.is_blocked ? 'Заблокировано' : 'Активно' }}
                </span>
              </p>
            </div>
            <div class="connection-info-section">
              <p><strong>Договор:</strong> {{ stats.connection.contract_number }}</p>
              <p><strong>Клиент:</strong> {{ stats.connection.client_name }}</p>
              <p><strong>Тариф:</strong> {{ stats.connection.tariff_name }}</p>
              <p><strong>Оборудование:</strong> {{ stats.connection.equipment_model }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Статистика трафика -->
      <div class="stats-grid mb-6">
        <div class="card stat-card">
          <div class="card-body">
            <h3 class="stat-label">Всего записей</h3>
            <p class="stat-value text-primary">{{ stats.traffic_stats.total_records }}</p>
          </div>
        </div>
        <div class="card stat-card">
          <div class="card-body">
            <h3 class="stat-label">Общий трафик</h3>
            <p class="stat-value text-success">{{ formatBytes(stats.traffic_stats.total_traffic) }}</p>
          </div>
        </div>
        <div class="card stat-card">
          <div class="card-body">
            <h3 class="stat-label">Активных дней</h3>
            <p class="stat-value text-primary">{{ stats.traffic_stats.active_days }}</p>
          </div>
        </div>
        <div class="card stat-card">
          <div class="card-body">
            <h3 class="stat-label">Средний трафик</h3>
            <p class="stat-value text-warning">{{ formatBytes(stats.traffic_stats.avg_traffic) }}</p>
          </div>
        </div>
      </div>

      <!-- Дополнительная статистика -->
      <div class="additional-stats-grid mb-6">
        <div class="card stat-card">
          <div class="card-body">
            <h3 class="stat-label">Входящий трафик</h3>
            <p class="stat-value text-success">{{ formatBytes(stats.traffic_stats.total_bytes_in) }}</p>
          </div>
        </div>
        <div class="card stat-card">
          <div class="card-body">
            <h3 class="stat-label">Исходящий трафик</h3>
            <p class="stat-value text-primary">{{ formatBytes(stats.traffic_stats.total_bytes_out) }}</p>
          </div>
        </div>
        <div class="card stat-card">
          <div class="card-body">
            <h3 class="stat-label">Максимальный трафик за сеанс</h3>
            <p class="stat-value text-danger">{{ formatBytes(stats.traffic_stats.max_traffic) }}</p>
          </div>
        </div>
      </div>

      <!-- Топ дней по трафику -->
      <div v-if="stats.top_days && stats.top_days.length > 0" class="card">
        <div class="card-header">
          <h2 class="text-xl font-semibold text-gray-900">Топ дней по трафику</h2>
        </div>
        <div class="table-container">
          <table class="table">
            <thead>
              <tr>
                <th>Дата</th>
                <th>Общий трафик</th>
                <th>День недели</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(day, index) in stats.top_days" :key="day.date">
                <td>
                  <div class="flex items-center gap-2">
                    <span class="font-medium">{{ formatDate(day.date) }}</span>
                    <span v-if="index === 0" class="badge badge-warning">
                      Пик
                    </span>
                  </div>
                </td>
                <td class="font-medium">
                  {{ formatBytes(day.daily_traffic) }}
                </td>
                <td class="text-gray-600">
                  {{ getDayOfWeek(day.date) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import apiClient from '../api/client';
import { formatDate } from '@/utils/dateUtils';

export default {
  name: 'ConnectionStats',
  setup() {
    const route = useRoute();
    const connectionId = route.params.id;
    
    const stats = ref(null);
    const loading = ref(false);
    const error = ref(null);

    const filters = reactive({
      fromDate: '',
      toDate: ''
    });

    const formatBytes = (bytes) => {
      if (bytes === 0 || bytes === null || bytes === undefined) return '0 B';
      const k = 1024;
      const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
      const i = Math.floor(Math.log(bytes) / Math.log(k));
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    const getDayOfWeek = (dateString) => {
      const days = ['Воскресенье', 'Понедельник', 'Вторник', 'Среда', 'Четверг', 'Пятница', 'Суббота'];
      const date = new Date(dateString);
      return days[date.getDay()];
    };

    const loadConnectionStats = async () => {
      if (!connectionId) {
        error.value = 'ID подключения не указан';
        return;
      }
      
      loading.value = true;
      error.value = null;
      
      try {
        const params = new URLSearchParams();
        if (filters.fromDate) params.append('from', filters.fromDate);
        if (filters.toDate) params.append('to', filters.toDate);
        
        const queryString = params.toString();
        const url = `/connections/${connectionId}/stats${queryString ? '?' + queryString : ''}`;
        
        const response = await apiClient.get(url);
        stats.value = response.data;
        
      } catch (e) {
        if (e.response?.status === 404) {
          error.value = 'Подключение не найдено';
        } else if (e.response) {
          error.value = `Ошибка при загрузке статистики: ${e.response.status}`;
        } else {
          error.value = 'Ошибка сети при загрузке статистики подключения';
        }
        console.error('Error loading connection stats:', e);
      } finally {
        loading.value = false;
      }
    };

    onMounted(() => {
      // Устанавливаем период по умолчанию - последние 30 дней
      const today = new Date();
      const thirtyDaysAgo = new Date(today);
      thirtyDaysAgo.setDate(today.getDate() - 30);
      
      filters.fromDate = thirtyDaysAgo.toISOString().split('T')[0];
      filters.toDate = today.toISOString().split('T')[0];
      
      loadConnectionStats();
    });

    return {
      stats,
      loading,
      error,
      filters,
      formatBytes,
      formatDate,
      getDayOfWeek,
      loadConnectionStats
    };
  }
};
</script>

<style scoped>
.connection-stats {
  max-width: 100%;
}

.dashboard-title {
  font-size: 2rem;
  font-weight: 700;
  color: var(--gray-900);
  margin-bottom: 2rem;
}

.card-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--gray-900);
  margin-bottom: 1.5rem;
}

.filter-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
  margin-bottom: 1.5rem;
}

.connection-info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
}

.connection-info-section p {
  margin-bottom: 0.75rem;
  line-height: 1.5;
}

.connection-info-section strong {
  color: var(--gray-700);
  font-weight: 600;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
}

.additional-stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
}

.stat-card {
  min-height: 100px;
}

.stat-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--gray-500);
  margin-bottom: 0.5rem;
}

.stat-value {
  font-size: 1.875rem;
  font-weight: 700;
  line-height: 1.2;
}

.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 3rem;
}

@media (max-width: 768px) {
  .connection-info-grid {
    grid-template-columns: 1fr;
  }
  
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .additional-stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>