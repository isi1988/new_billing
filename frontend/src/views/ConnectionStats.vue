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

      <!-- Детальные потоки трафика -->
      <div class="card mb-6">
        <div class="card-header flex justify-between items-center">
          <h2 class="text-xl font-semibold text-gray-900">Детальные потоки трафика</h2>
          <div class="text-sm text-gray-500">
            Показано {{ flows.length }} из {{ flowsPagination.total }} записей
          </div>
        </div>

        <!-- Загрузка flows -->
        <div v-if="flowsLoading" class="loading-container">
          <div class="spinner"></div>
        </div>

        <!-- Таблица flows -->
        <div v-else class="table-container">
          <table class="table">
            <thead>
              <tr>
                <th>Время</th>
                <th>Источник</th>
                <th>Назначение</th>
                <th>Протокол</th>
                <th>Направление</th>
                <th>Трафик</th>
                <th>Пакеты</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="flows.length === 0">
                <td colspan="7" class="text-center text-gray-500">Нет данных для отображения</td>
              </tr>
              <tr v-for="item in flows" :key="item.timestamp + item.src_ip + item.dst_ip">
                <td class="text-gray-900">{{ formatDateTime(item.timestamp) }}</td>
                <td class="text-gray-700">
                  <div class="flow-endpoint">
                    <IPDisplay :ip="item.src_ip" />
                    <div class="text-xs text-gray-500 mt-1">Порт: {{ item.src_port }}</div>
                  </div>
                </td>
                <td class="text-gray-700">
                  <div class="flow-endpoint">
                    <IPDisplay :ip="item.dst_ip" />
                    <div class="text-xs text-gray-500 mt-1">Порт: {{ item.dst_port }}</div>
                  </div>
                </td>
                <td class="text-gray-600">{{ getProtocolName(item.protocol) }}</td>
                <td>
                  <span :class="getDirectionClass(item.direction)">
                    {{ getDirectionLabel(item.direction) }}
                  </span>
                </td>
                <td class="font-medium">{{ formatBytes(item.bytes) }}</td>
                <td class="text-gray-600">{{ item.packets }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Пагинация flows -->
        <div class="card-footer flex justify-between items-center">
          <div class="flex items-center gap-2">
            <label class="form-label">Записей на странице:</label>
            <select 
              v-model="flowsPagination.limit" 
              @change="loadFlows"
              class="form-control"
              style="width: auto;"
            >
              <option value="10">10</option>
              <option value="25">25</option>
              <option value="50">50</option>
              <option value="100">100</option>
            </select>
          </div>
          <div class="pagination-controls">
            <button 
              @click="previousFlowsPage"
              :disabled="flowsPagination.offset === 0"
              class="btn btn-sm btn-secondary"
            >
              <span class="material-icons icon-xs">arrow_back</span>
              Назад
            </button>
            <span class="pagination-info">
              Страница {{ currentFlowsPage }} из {{ totalFlowsPages }}
            </span>
            <button 
              @click="nextFlowsPage"
              :disabled="!hasNextFlowsPage"
              class="btn btn-sm btn-secondary"
            >
              Вперед
              <span class="material-icons icon-xs">arrow_forward</span>
            </button>
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
import { ref, reactive, computed, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import apiClient from '../api/client';
import { formatDate } from '@/utils/dateUtils';
import IPDisplay from '../components/ui/IPDisplay.vue';

export default {
  name: 'ConnectionStats',
  components: {
    IPDisplay
  },
  setup() {
    const route = useRoute();
    const connectionId = route.params.id;
    
    const stats = ref(null);
    const loading = ref(false);
    const error = ref(null);

    // Flows data
    const flows = ref([]);
    const flowsLoading = ref(false);
    const flowsPagination = reactive({
      limit: 25,
      offset: 0,
      total: 0
    });

    const filters = reactive({
      fromDate: '',
      toDate: ''
    });

    // Computed properties for flows pagination
    const currentFlowsPage = computed(() => Math.floor(flowsPagination.offset / flowsPagination.limit) + 1);
    const totalFlowsPages = computed(() => Math.ceil(flowsPagination.total / flowsPagination.limit));
    const hasNextFlowsPage = computed(() => flowsPagination.offset + flowsPagination.limit < flowsPagination.total);

    const formatBytes = (bytes) => {
      if (bytes === 0 || bytes === null || bytes === undefined) return '0 B';
      const k = 1024;
      const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
      const i = Math.floor(Math.log(bytes) / Math.log(k));
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    const formatDateTime = (dateString) => {
      if (!dateString) return 'N/A';
      return new Date(dateString).toLocaleString('ru-RU', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      });
    };

    const getDayOfWeek = (dateString) => {
      const days = ['Воскресенье', 'Понедельник', 'Вторник', 'Среда', 'Четверг', 'Пятница', 'Суббота'];
      const date = new Date(dateString);
      return days[date.getDay()];
    };

    // Flow helper functions
    const getProtocolName = (protocolNumber) => {
      const protocols = {
        1: 'ICMP',
        6: 'TCP',
        17: 'UDP',
        47: 'GRE',
        50: 'ESP',
        51: 'AH'
      };
      return protocols[protocolNumber] || `Protocol ${protocolNumber}`;
    };

    const getDirectionLabel = (direction) => {
      switch (direction) {
        case 'incoming': return 'Входящий';
        case 'outgoing': return 'Исходящий';
        case 'mixed': return 'Смешанный';
        default: return 'Неизвестно';
      }
    };

    const getDirectionClass = (direction) => {
      switch (direction) {
        case 'incoming': return 'direction-badge incoming';
        case 'outgoing': return 'direction-badge outgoing';
        case 'mixed': return 'direction-badge mixed';
        default: return 'direction-badge';
      }
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
        
        // Загружаем flows после загрузки статистики
        await loadFlows();
        
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

    const loadFlows = async () => {
      if (!connectionId || !stats.value) {
        return;
      }
      
      flowsLoading.value = true;
      
      try {
        // Используем IP адрес из статистики подключения
        const searchIP = stats.value.connection.ip_address;
        
        const params = new URLSearchParams({
          ip: searchIP,
          page: Math.floor(flowsPagination.offset / flowsPagination.limit) + 1,
          limit: flowsPagination.limit
        });
        
        if (filters.fromDate) params.append('from', filters.fromDate);
        if (filters.toDate) params.append('to', filters.toDate);
        
        console.log('Loading flows with params:', params.toString());
        
        const response = await apiClient.get(`/flows/search?${params.toString()}`);
        const result = response.data;
        
        flows.value = result.flows || [];
        flowsPagination.total = result.total_records || 0;
        
      } catch (e) {
        console.error('Error loading flows:', e);
        flows.value = [];
        flowsPagination.total = 0;
      } finally {
        flowsLoading.value = false;
      }
    };

    const nextFlowsPage = () => {
      if (hasNextFlowsPage.value) {
        flowsPagination.offset += flowsPagination.limit;
        loadFlows();
      }
    };

    const previousFlowsPage = () => {
      if (flowsPagination.offset > 0) {
        flowsPagination.offset = Math.max(0, flowsPagination.offset - flowsPagination.limit);
        loadFlows();
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
      flows,
      flowsLoading,
      flowsPagination,
      currentFlowsPage,
      totalFlowsPages,
      hasNextFlowsPage,
      formatBytes,
      formatDate,
      formatDateTime,
      getDayOfWeek,
      getProtocolName,
      getDirectionLabel,
      getDirectionClass,
      loadConnectionStats,
      loadFlows,
      nextFlowsPage,
      previousFlowsPage
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

/* Стили для flows таблицы */
.flow-endpoint {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.direction-badge {
  padding: 0.25rem 0.5rem;
  border-radius: 0.375rem;
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: uppercase;
}

.direction-badge.incoming {
  background-color: #dbeafe;
  color: #1e40af;
}

.direction-badge.outgoing {
  background-color: #fef3c7;
  color: #92400e;
}

.direction-badge.mixed {
  background-color: #f3e8ff;
  color: #7c3aed;
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.pagination-info {
  font-size: 0.875rem;
  color: var(--gray-600);
  padding: 0.5rem 1rem;
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
  
  .card-footer {
    flex-direction: column;
    gap: 1rem;
    align-items: stretch;
  }
  
  .pagination-controls {
    justify-content: center;
  }
}
</style>