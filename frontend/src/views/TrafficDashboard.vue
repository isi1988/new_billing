<template>
  <div class="traffic-dashboard">
    <h1 class="dashboard-title">Дашборд трафика</h1>
    
    <!-- Фильтры поиска -->
    <div class="card mb-6">
      <div class="card-body">
        <h2 class="card-title">Фильтры</h2>
        <div class="filter-grid">
          <div class="form-group">
            <label class="form-label">Клиент</label>
            <div class="client-search-container">
              <input 
                v-model="clientSearchQuery"
                type="text"
                class="form-control"
                placeholder="Поиск клиента по имени..."
                @input="searchClients"
                @focus="showClientDropdown = true"
                @blur="hideClientDropdown"
              />
              <div v-if="showClientDropdown && filteredClients.length > 0" class="client-dropdown">
                <div 
                  v-for="client in filteredClients" 
                  :key="client.id"
                  class="client-option"
                  @mousedown="selectClient(client)"
                >
                  <div class="client-name">
                    {{ getClientDisplayName(client) }}
                  </div>
                  <div class="client-email">{{ client.email }}</div>
                </div>
              </div>
            </div>
          </div>
          <div class="form-group">
            <label class="form-label">IP Адрес или сеть</label>
            <input 
              v-model="filters.ipAddress"
              type="text"
              class="form-control"
              placeholder="192.168.1.1 или 192.168.1.0/24 или 192.168.1.*"
            />
            <small class="form-text">
              Поддерживается: точный IP, CIDR (192.168.1.0/24), маска с * (192.168.1.*)
            </small>
          </div>
          <div class="form-group">
            <label class="form-label">Дата с</label>
            <input 
              v-model="filters.fromDate"
              type="datetime-local"
              class="form-control"
            />
          </div>
          <div class="form-group">
            <label class="form-label">Дата по</label>
            <input 
              v-model="filters.toDate"
              type="datetime-local"
              class="form-control"
            />
          </div>
        </div>
        <div class="button-group">
          <button 
            @click="searchTraffic"
            class="btn btn-md btn-primary"
          >
            <span class="material-icons icon-sm">search</span>
            Найти
          </button>
          <button 
            @click="clearFilters"
            class="btn btn-md btn-secondary"
          >
            <span class="material-icons icon-sm">clear</span>
            Очистить
          </button>
          <button 
            @click="exportToCSV"
            :disabled="loading"
            class="btn btn-md btn-success"
          >
            <span class="material-icons icon-sm">file_download</span>
            {{ loading ? 'Экспорт...' : 'Экспорт CSV' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Статистика -->
    <div v-if="stats" class="stats-grid mb-6">
      <div class="card stat-card">
        <div class="card-body">
          <h3 class="stat-label">Всего записей</h3>
          <p class="stat-value text-primary">{{ stats.total_records }}</p>
        </div>
      </div>
      <div class="card stat-card">
        <div class="card-body">
          <h3 class="stat-label">Входящий трафик</h3>
          <p class="stat-value text-success">{{ formatBytes(stats.total_bytes_in) }}</p>
        </div>
      </div>
      <div class="card stat-card">
        <div class="card-body">
          <h3 class="stat-label">Исходящий трафик</h3>
          <p class="stat-value text-info">{{ formatBytes(stats.total_bytes_out) }}</p>
        </div>
      </div>
      <div class="card stat-card">
        <div class="card-body">
          <h3 class="stat-label">Общий трафик</h3>
          <p class="stat-value text-warning">{{ formatBytes(stats.total_traffic) }}</p>
        </div>
      </div>
    </div>

    <!-- Таблица трафика -->
    <div class="card">
      <div class="card-header flex justify-between items-center">
        <h2 class="text-xl font-semibold text-gray-900">Данные трафика</h2>
        <div class="text-sm text-gray-500">
          Показано {{ traffic.length }} из {{ pagination.total }} записей
        </div>
      </div>

      <!-- Загрузка -->
      <div v-if="loading" class="loading-container">
        <div class="spinner"></div>
      </div>

      <!-- Ошибка -->
      <div v-else-if="error" class="error-container">
        <p class="text-danger">{{ error }}</p>
      </div>

      <!-- Таблица -->
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
            <tr v-if="traffic.length === 0">
              <td colspan="7" class="text-center text-gray-500">Нет данных для отображения</td>
            </tr>
            <tr v-for="item in traffic" :key="item.timestamp + item.src_ip + item.dst_ip">
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
                <span :class="getDirectionClass(item)">
                  {{ getDirectionLabel(item) }}
                </span>
              </td>
              <td class="font-medium">{{ formatBytes(item.bytes) }}</td>
              <td class="text-gray-600">{{ item.packets }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Пагинация -->
      <div class="card-footer flex justify-between items-center">
        <div class="flex items-center gap-2">
          <label class="form-label">Записей на странице:</label>
          <select 
            v-model="pagination.limit" 
            @change="searchTraffic"
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
            @click="previousPage"
            :disabled="pagination.offset === 0"
            class="btn btn-sm btn-secondary"
          >
            <span class="material-icons icon-xs">arrow_back</span>
            Назад
          </button>
          <span class="pagination-info">
            Страница {{ currentPage }} из {{ totalPages }}
          </span>
          <button 
            @click="nextPage"
            :disabled="!hasNextPage"
            class="btn btn-sm btn-secondary"
          >
            Вперед
            <span class="material-icons icon-xs">arrow_forward</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, reactive, computed, onMounted } from 'vue';
import apiClient from '../api/client';
import IPDisplay from '../components/ui/IPDisplay.vue';

export default {
  name: 'TrafficDashboard',
  components: {
    IPDisplay
  },
  setup() {
    const traffic = ref([]);
    const stats = ref(null);
    const loading = ref(false);
    const error = ref(null);
    
    // Client search functionality
    const clients = ref([]);
    const clientSearchQuery = ref('');
    const showClientDropdown = ref(false);
    const filteredClients = ref([]);
    const selectedClient = ref(null);
    const currentSearchIP = ref('');
    const currentSearchMask = ref('');
    const clientConnectionIPs = ref([]);

    const filters = reactive({
      clientId: '',
      ipAddress: '',
      fromDate: '',
      toDate: ''
    });

    const pagination = reactive({
      limit: 25,
      offset: 0,
      total: 0
    });

    const currentPage = computed(() => Math.floor(pagination.offset / pagination.limit) + 1);
    const totalPages = computed(() => Math.ceil(pagination.total / pagination.limit));
    const hasNextPage = computed(() => pagination.offset + pagination.limit < pagination.total);

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

    // Вспомогательные функции для flows
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

    const getDirectionLabel = (item) => {
      // Используем направление из API ответа
      if (item.direction === 'incoming') {
        return 'Входящий';
      }
      if (item.direction === 'outgoing') {
        return 'Исходящий';
      }
      if (item.direction === 'mixed') {
        return 'Смешанный';
      }
      return 'Неизвестно';
    };

    const getDirectionClass = (item) => {
      // Используем направление из API ответа
      if (item.direction === 'incoming') {
        return 'direction-badge incoming';
      }
      if (item.direction === 'outgoing') {
        return 'direction-badge outgoing';
      }
      if (item.direction === 'mixed') {
        return 'direction-badge mixed';
      }
      return 'direction-badge';
    };

    const buildQueryParams = () => {
      const params = new URLSearchParams();
      
      if (filters.clientId) params.append('client_id', filters.clientId);
      if (filters.ipAddress) params.append('ip_address', filters.ipAddress);
      if (filters.fromDate) params.append('from', filters.fromDate.replace('T', ' '));
      if (filters.toDate) params.append('to', filters.toDate.replace('T', ' '));
      
      params.append('limit', pagination.limit);
      params.append('offset', pagination.offset);
      
      return params.toString();
    };

    const searchTraffic = async () => {
      loading.value = true;
      error.value = null;
      
      try {
        // Проверяем, что есть хотя бы один критерий поиска
        if (!selectedClient.value && !filters.ipAddress) {
          error.value = 'Необходимо указать IP-адрес или выбрать клиента';
          loading.value = false;
          return;
        }

        // Определяем searchIP - либо из selectedClient, либо из filters.ipAddress
        let searchIP = filters.ipAddress;
        if (selectedClient.value && !searchIP) {
          // Если выбран клиент, но нет IP, получаем IP клиента
          try {
            const clientResponse = await apiClient.get(`/clients/${selectedClient.value.id}/connections`);
            const connections = clientResponse.data || [];
            if (connections.length > 0) {
              // Сохраняем все IP подключений клиента для определения направления
              clientConnectionIPs.value = connections.map(conn => conn.ip_address);
              searchIP = connections[0].ip_address;
              currentSearchMask.value = connections[0].mask;
            } else {
              error.value = 'У выбранного клиента нет подключений';
              loading.value = false;
              return;
            }
          } catch (e) {
            error.value = 'Ошибка при получении IP адресов клиента';
            loading.value = false;
            return;
          }
        } else {
          // Если поиск по конкретному IP, очищаем массив подключений клиента
          clientConnectionIPs.value = [];
          currentSearchMask.value = '';
        }

        // Сохраняем текущий IP для определения направления трафика
        currentSearchIP.value = searchIP;

        // Формируем параметры запроса
        const params = new URLSearchParams({
          ip: searchIP,
          page: Math.floor(pagination.offset / pagination.limit) + 1,
          limit: pagination.limit
        });
        
        // Добавляем маску, если есть
        if (currentSearchMask.value) {
          params.append('mask', currentSearchMask.value);
        }
        
        if (filters.fromDate) params.append('from', filters.fromDate.split('T')[0]);
        if (filters.toDate) params.append('to', filters.toDate.split('T')[0]);
        
        console.log('Searching flows with params:', params.toString());
        
        // Получаем flows данные
        const response = await apiClient.get(`/flows/search?${params.toString()}`);
        console.log('Flows response:', response.data);
        
        const result = response.data;
        traffic.value = result.flows || [];
        
        // Обновляем статистику и пагинацию из ответа
        stats.value = {
          total_records: result.total_records,
          total_bytes_in: result.total_bytes_in,
          total_bytes_out: result.total_bytes_out,
          total_traffic: result.total_traffic
        };
        
        pagination.total = result.total_records;
        
      } catch (e) {
        error.value = 'Ошибка при загрузке данных трафика';
        console.error('Search traffic error:', e);
        traffic.value = [];
        stats.value = null;
      } finally {
        loading.value = false;
      }
    };

    const clearFilters = () => {
      filters.clientId = '';
      filters.ipAddress = '';
      filters.fromDate = '';
      filters.toDate = '';
      clientSearchQuery.value = '';
      selectedClient.value = null;
      filteredClients.value = clients.value;
      pagination.offset = 0;
      traffic.value = [];
      stats.value = null;
      error.value = null;
    };

    const exportToCSV = async () => {
      // Проверяем, что есть критерии поиска для экспорта
      if (!selectedClient.value && !filters.ipAddress) {
        error.value = 'Необходимо указать IP-адрес или выбрать клиента для экспорта';
        return;
      }
      
      loading.value = true;
      error.value = null;
      
      try {
        // Определяем searchIP аналогично searchTraffic
        let searchIP = filters.ipAddress;
        let searchMask = '';
        if (selectedClient.value && !searchIP) {
          const clientResponse = await apiClient.get(`/clients/${selectedClient.value.id}/connections`);
          const connections = clientResponse.data || [];
          if (connections.length > 0) {
            searchIP = connections[0].ip_address;
            searchMask = connections[0].mask;
          } else {
            error.value = 'У выбранного клиента нет подключений';
            loading.value = false;
            return;
          }
        }

        // Формируем параметры для экспорта без пагинации
        const params = new URLSearchParams({
          ip: searchIP
        });
        
        // Добавляем маску, если есть
        if (searchMask) {
          params.append('mask', searchMask);
        }
        
        if (filters.fromDate) params.append('from', filters.fromDate.split('T')[0]);
        if (filters.toDate) params.append('to', filters.toDate.split('T')[0]);
        
        console.log('Exporting CSV with params:', params.toString());
        
        const response = await apiClient.get(`/flows/export?${params.toString()}`, {
          responseType: 'blob'
        });
        
        // Создаем ссылку для скачивания
        const url = window.URL.createObjectURL(new Blob([response.data]));
        const link = document.createElement('a');
        link.href = url;
        link.setAttribute('download', `flows_export_${new Date().toISOString().slice(0, 19).replace(/:/g, '-')}.csv`);
        document.body.appendChild(link);
        link.click();
        link.remove();
        window.URL.revokeObjectURL(url);
        
        console.log('CSV export completed successfully');
        
      } catch (e) {
        error.value = 'Ошибка при экспорте данных';
        console.error('Export error:', e);
      } finally {
        loading.value = false;
      }
    };

    const nextPage = () => {
      if (hasNextPage.value) {
        pagination.offset += pagination.limit;
        searchTraffic();
      }
    };

    const previousPage = () => {
      if (pagination.offset > 0) {
        pagination.offset = Math.max(0, pagination.offset - pagination.limit);
        searchTraffic();
      }
    };

    // Client search functions
    const loadClients = async () => {
      try {
        const response = await apiClient.get('/clients');
        clients.value = response.data || [];
        filteredClients.value = clients.value;
      } catch (e) {
        console.error('Ошибка при загрузке клиентов:', e);
      }
    };

    const getClientDisplayName = (client) => {
      if (client.client_type === 'individual') {
        return `${client.last_name || ''} ${client.first_name || ''}`.trim() || client.email;
      }
      return client.short_name || client.full_name || client.email;
    };

    const searchClients = () => {
      if (!clientSearchQuery.value.trim()) {
        filteredClients.value = clients.value;
        return;
      }
      
      const query = clientSearchQuery.value.toLowerCase();
      filteredClients.value = clients.value.filter(client => {
        const displayName = getClientDisplayName(client).toLowerCase();
        const email = (client.email || '').toLowerCase();
        return displayName.includes(query) || email.includes(query);
      });
    };

    const selectClient = (client) => {
      selectedClient.value = client;
      clientSearchQuery.value = getClientDisplayName(client);
      filters.clientId = client.id;
      showClientDropdown.value = false;
    };

    const hideClientDropdown = () => {
      setTimeout(() => {
        showClientDropdown.value = false;
      }, 200);
    };

    onMounted(() => {
      loadClients();
      // Не вызываем searchTraffic() автоматически, так как нужны критерии поиска
    });

    return {
      traffic,
      stats,
      loading,
      error,
      filters,
      pagination,
      currentPage,
      totalPages,
      hasNextPage,
      formatBytes,
      formatDateTime,
      searchTraffic,
      clearFilters,
      exportToCSV,
      nextPage,
      previousPage,
      // Client search
      clients,
      clientSearchQuery,
      showClientDropdown,
      filteredClients,
      selectedClient,
      currentSearchIP,
      currentSearchMask,
      clientConnectionIPs,
      getClientDisplayName,
      searchClients,
      selectClient,
      hideClientDropdown,
      // Flow helpers
      getProtocolName,
      getDirectionLabel,
      getDirectionClass
    };
  }
};
</script>

<style scoped>
.traffic-dashboard {
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

.button-group {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
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

.error-container {
  text-align: center;
  padding: 3rem;
}

.table-container {
  overflow-x: auto;
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

/* Client Search Dropdown */
.client-search-container {
  position: relative;
}

.client-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: white;
  border: 1px solid var(--gray-300);
  border-radius: var(--radius-lg);
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  max-height: 200px;
  overflow-y: auto;
  z-index: 9999;
  margin-top: 2px;
}

.client-option {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--gray-100);
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.client-option:hover {
  background-color: var(--primary-50);
}

.client-option:last-child {
  border-bottom: none;
}

.client-name {
  font-weight: 500;
  color: var(--gray-900);
  margin-bottom: 0.25rem;
}

.client-email {
  font-size: 0.875rem;
  color: var(--gray-600);
}

@media (max-width: 768px) {
  .filter-grid {
    grid-template-columns: 1fr;
  }
  
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .button-group {
    flex-direction: column;
  }
  
  .card-footer {
    flex-direction: column;
    gap: 1rem;
    align-items: stretch;
  }
  
  .pagination-controls {
    justify-content: center;
  }
  
  .client-dropdown {
    position: fixed;
    left: 1rem;
    right: 1rem;
    max-height: 50vh;
  }
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

.form-text {
  font-size: 0.75rem;
  color: var(--gray-600);
  margin-top: 0.25rem;
  display: block;
}
</style>