<script setup>
import { ref, computed, reactive, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useCrud } from '@/composables/useCrud';
import DataTable from '@/components/ui/DataTable.vue';
import Modal from '@/components/ui/Modal.vue';
import ContractForm from '@/components/forms/ContractForm.vue';
import ConnectionForm from '@/components/forms/ConnectionForm.vue';
import StatusBadge from '@/components/ui/StatusBadge.vue';
import SearchFilters from '@/components/ui/SearchFilters.vue';
import apiClient from '@/api/client';
import { formatDate } from '@/utils/dateUtils';
import { useNotificationStore } from '@/stores/notification';

const notificationStore = useNotificationStore();

// Инициализируем роутер
const router = useRouter();

// Инициализируем CRUD-операции для эндпоинта '/api/contracts'
const {
  items: contracts,
  loading,
  createItem,
  updateItem,
  deleteItem
} = useCrud('contracts');

// Состояние для управления модальным окном договоров
const isModalOpen = ref(false);
const currentContract = ref(null);
const isEditMode = ref(false);

// Состояние для управления модальным окном подключений
const showConnectionModal = ref(false);
const currentConnection = ref(null);
const isConnectionEditMode = ref(false);

// Состояние для расширяемых подключений
const expandedContracts = ref(new Set());
const contractConnections = ref({});
const loadingConnections = ref(new Set());

// Search and filter state
const searchQuery = ref('');
const filterValues = reactive({
  is_blocked: '',
  client_id: ''
});

// Filter configuration
const filters = [
  {
    key: 'is_blocked',
    label: 'Статус',
    type: 'select',
    options: [
      { value: 'false', label: 'Активные' },
      { value: 'true', label: 'Заблокированные' }
    ]
  },
  {
    key: 'client_id',
    label: 'ID клиента',
    type: 'number',
    placeholder: 'Введите ID клиента...'
  }
];

// Computed filtered contracts
const filteredContracts = computed(() => {
  let filtered = contracts.value;

  // Apply search
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(contract => {
      const number = (contract.number || '').toLowerCase();
      const clientName = (contract.client_name || '').toLowerCase();
      const clientEmail = (contract.client_email || '').toLowerCase();
      return number.includes(query) || clientName.includes(query) || clientEmail.includes(query);
    });
  }

  // Apply filters
  if (filterValues.is_blocked !== '') {
    const isBlocked = filterValues.is_blocked === 'true';
    filtered = filtered.filter(contract => contract.is_blocked === isBlocked);
  }
  if (filterValues.client_id) {
    filtered = filtered.filter(contract => contract.client_id.toString() === filterValues.client_id.toString());
  }

  return filtered;
});

// Описание колонок для таблицы
const columns = [
  { key: 'id', label: 'ID' },
  { key: 'number', label: 'Номер' },
  { key: 'sign_date', label: 'Дата подписания', formatter: (contract) => formatDate(contract.sign_date) },
  { 
    key: 'client_name', 
    label: 'Клиент', 
    formatter: (contract) => {
      const name = contract.client_name || `Клиент ID: ${contract.client_id}`;
      return name;
    }
  },
  { key: 'connections_count', label: 'Подключений' },
  {
    key: 'is_blocked',
    label: 'Статус',
    component: 'StatusBadge'
  },
];


// Открытие модального окна для создания нового договора
function openCreateModal() {
  isEditMode.value = false;
  // Данные по умолчанию: сегодняшняя дата
  currentContract.value = {
    number: '',
    sign_date: new Date().toISOString().split('T')[0],
    client_id: null,
  };
  isModalOpen.value = true;
}

// Открытие модального окна для редактирования
function openEditModal(item) {
  isEditMode.value = true;
  currentContract.value = { ...item };
  isModalOpen.value = true;
}

// Обработка сохранения данных из формы
async function handleSave(contractData) {
  try {
    if (isEditMode.value) {
      await updateItem(contractData.id, contractData);
      notificationStore.addNotification({
        type: 'success',
        title: 'Договор обновлён',
        message: 'Данные договора успешно обновлены'
      });
    } else {
      await createItem(contractData);
      notificationStore.addNotification({
        type: 'success',
        title: 'Договор создан',
        message: 'Новый договор успешно создан'
      });
    }
    isModalOpen.value = false;
    currentContract.value = null; // Очищаем форму
  } catch (error) {
    notificationStore.addNotification({
      type: 'error',
      title: 'Ошибка сохранения',
      message: 'Не удалось сохранить договор'
    });
  }
}

// Обработка сохранения подключения
async function handleConnectionSave(connectionData) {
  try {
    if (isConnectionEditMode.value) {
      await apiClient.put(`/connections/${connectionData.id}`, connectionData);
    } else {
      await apiClient.post('/connections', connectionData);
    }
    showConnectionModal.value = false;
    currentConnection.value = null;
    
    // Обновляем список подключений для соответствующего договора
    if (connectionData.contract_id) {
      await refreshConnections(connectionData.contract_id);
    }
  } catch (error) {
    console.error('Ошибка сохранения подключения:', error);
    notificationStore.addNotification({
      type: 'error',
      title: 'Ошибка сохранения',
      message: 'Не удалось сохранить подключение'
    });
  }
}

// Обработка удаления
async function handleDelete(itemId) {
  {
    try {
      await deleteItem(itemId);
      notificationStore.addNotification({
        type: 'success',
        title: 'Договор удалён',
        message: 'Договор успешно удалён'
      });
    } catch (error) {
      notificationStore.addNotification({
        type: 'error',
        title: 'Ошибка удаления',
        message: 'Не удалось удалить договор'
      });
    }
  }
}

async function handleBlock(contract) {
  const action = contract.is_blocked ? 'разблокировать' : 'заблокировать';
  {
    try {
      const endpoint = contract.is_blocked ? 'unblock' : 'block';
      await apiClient.post(`/contracts/${contract.id}/${endpoint}`);
      
      // Обновляем статус договора в локальном массиве
      const index = contracts.value.findIndex(c => c.id === contract.id);
      if (index !== -1) {
        contracts.value[index].is_blocked = !contract.is_blocked;
      }
    } catch (error) {
      notificationStore.addNotification({
        type: 'error',
        title: 'Ошибка операции',
        message: `Не удалось ${action} договор`
      });
    }
  }
}

// Функция для переключения состояния расширения подключений
async function toggleConnections(contract) {
  if (expandedContracts.value.has(contract.id)) {
    expandedContracts.value.delete(contract.id);
  } else {
    expandedContracts.value.add(contract.id);
    
    if (!contractConnections.value[contract.id]) {
      loadingConnections.value.add(contract.id);
      
      try {
        const response = await apiClient.get(`/contracts/${contract.id}/connections`);
        contractConnections.value[contract.id] = response.data || [];
      } catch (error) {
        console.error('Ошибка загрузки подключений:', error);
        notificationStore.addNotification({
          type: 'error',
          title: 'Ошибка загрузки',
          message: 'Не удалось загрузить подключения'
        });
        contractConnections.value[contract.id] = [];
      } finally {
        loadingConnections.value.delete(contract.id);
      }
    }
  }
}

// Функция блокировки/разблокировки подключения
async function handleConnectionBlock(connection, contractId) {
  const action = connection.is_blocked ? 'разблокировать' : 'заблокировать';
  try {
    const endpoint = connection.is_blocked ? 'unblock' : 'block';
    await apiClient.post(`/connections/${connection.id}/${endpoint}`);
    
    // Обновляем статус подключения в локальном массиве
    if (contractConnections.value[contractId]) {
      const index = contractConnections.value[contractId].findIndex(c => c.id === connection.id);
      if (index !== -1) {
        contractConnections.value[contractId][index].is_blocked = !connection.is_blocked;
      }
    }
    
    notificationStore.addNotification({
      type: 'success',
      title: 'Операция выполнена',
      message: `Подключение успешно ${connection.is_blocked ? 'разблокировано' : 'заблокировано'}`
    });
  } catch (error) {
    notificationStore.addNotification({
      type: 'error',
      title: 'Ошибка операции',
      message: `Не удалось ${action} подключение`
    });
  }
}

// Функции для управления подключениями
function createConnection(contract) {
  // Открываем модальное окно создания подключения с предзаполненным contract_id
  currentConnection.value = {
    contract_id: contract.id,
    address: '',
    connection_type: '',
    ip_address: '',
    mask: 24,
    equipment_id: null,
    tariff_id: null,
    is_blocked: false
  };
  isConnectionEditMode.value = false;
  showConnectionModal.value = true;
}

function editConnection(connection) {
  // Переходим к странице редактирования подключения
  router.push({
    path: '/connections',
    query: {
      edit: connection.id,
      return_to: 'contracts'
    }
  });
}

// Функция для обновления подключений после изменений
async function refreshConnections(contractId) {
  if (expandedContracts.value.has(contractId)) {
    loadingConnections.value.add(contractId);
    
    try {
      const response = await apiClient.get(`/contracts/${contractId}/connections`);
      contractConnections.value[contractId] = response.data || [];
    } catch (error) {
      console.error('Ошибка загрузки подключений:', error);
    } finally {
      loadingConnections.value.delete(contractId);
    }
  }
}

// Обработка возврата со страницы подключений
onMounted(() => {
  const query = router.currentRoute.value.query;
  if (query.refreshConnections && query.contractId) {
    const contractId = parseInt(query.contractId);
    refreshConnections(contractId);
    // Очищаем query параметры
    router.replace({ query: {} });
  }
});

// Search and filter functions
function clearFilters() {
  searchQuery.value = '';
  filterValues.is_blocked = '';
  filterValues.client_id = '';
}
</script>

<template>
  <div class="page-container">
    <header class="page-header">
      <h1>Управление договорами</h1>
    </header>

    <SearchFilters
      :search-query="searchQuery"
      search-placeholder="Поиск по номеру договора или имени клиента..."
      :filters="filters"
      :filter-values="filterValues"
      @search="searchQuery = $event"
      @filter="filterValues[$event.key] = $event.value"
      @clear="clearFilters"
    />

    <DataTable
        :items="filteredContracts"
        :columns="columns"
        :loading="loading"
        @edit="openEditModal"
        @delete="handleDelete"
    >
      <template #cell-is_blocked="{ item }">
        <StatusBadge type="blocked_status" :value="item.is_blocked" size="small" />
      </template>

      <template #cell-connections_count="{ item }">
        <button 
          @click="toggleConnections(item)"
          class="connections-toggle"
          :class="{ 'expanded': expandedContracts.has(item.id) }"
        >
          <span class="count">{{ item.connections_count || '0' }}</span>
          <span class="material-icons icon-sm toggle-icon">{{ expandedContracts.has(item.id) ? 'expand_less' : 'expand_more' }}</span>
        </button>
      </template>
      
      <template #actions="{ item }">
        <router-link 
          :to="`/contracts/${item.id}/stats`"
          class="btn btn-icon btn-sm stats-btn mr-2"
          title="Статистика по договору"
        >
          <span class="material-icons icon-sm">analytics</span>
        </router-link>
        <button 
          @click="handleBlock(item)" 
          :class="['btn btn-icon btn-sm', item.is_blocked ? 'unblock-btn' : 'block-btn']"
          :title="item.is_blocked ? 'Разблокировать договор' : 'Заблокировать договор'"
        >
          <span class="material-icons icon-sm">{{ item.is_blocked ? 'lock_open' : 'lock' }}</span>
        </button>
      </template>

      <!-- Expandable rows for connections -->
      <template v-for="item in filteredContracts" :key="item.id" #[`expand-${item.id}`]>
        <div v-if="expandedContracts.has(item.id)" class="connections-expanded">
          <div v-if="loadingConnections.has(item.id)" class="loading">
            <span class="material-icons icon-sm">hourglass_empty</span>
            Загрузка подключений...
          </div>
          <div v-else-if="contractConnections[item.id]?.length === 0" class="no-connections">
            <p>Нет подключений</p>
            <button @click.stop="createConnection(item)" class="btn btn-sm btn-primary">
              <span class="material-icons icon-sm">add</span> Добавить
            </button>
          </div>
          <div v-else class="connections-list">
            <div class="connections-header">
              <strong>Подключения ({{ contractConnections[item.id]?.length || 0 }})</strong>
              <button @click.stop="createConnection(item)" class="btn btn-sm btn-primary">
                <span class="material-icons icon-sm">add</span> Добавить
              </button>
            </div>
            <div v-for="connection in contractConnections[item.id]" :key="connection.id" class="connection-item">
              <div class="connection-info">
                <div class="connection-main">
                  <span class="address">{{ connection.address }}</span>
                  <span class="ip">{{ connection.ip_address }}</span>
                  <StatusBadge type="blocked_status" :value="connection.is_blocked" size="small" />
                </div>
                <div class="connection-details">
                  <span class="detail">{{ connection.tariff_name || `Тариф ID: ${connection.tariff_id}` }}</span>
                  <span class="detail">{{ connection.equipment_model || `Оборудование ID: ${connection.equipment_id}` }}</span>
                </div>
              </div>
              <div class="connection-actions">
                <button 
                  @click.stop="editConnection(connection)"
                  class="btn btn-icon btn-sm edit-btn"
                  title="Редактировать"
                >
                  <span class="material-icons icon-sm">edit</span>
                </button>
                <button 
                  @click.stop="handleConnectionBlock(connection, item.id)" 
                  :class="['btn btn-icon btn-sm', connection.is_blocked ? 'unblock-btn' : 'block-btn']"
                  :title="connection.is_blocked ? 'Разблокировать' : 'Заблокировать'"
                >
                  <span class="material-icons icon-sm">{{ connection.is_blocked ? 'lock_open' : 'lock' }}</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </template>
    </DataTable>

    <button class="fab" @click="openCreateModal">
      <span class="material-icons icon-lg">add</span>
    </button>

    <Modal :is-open="isModalOpen" @close="isModalOpen = false">
      <template #header>
        <h2>{{ isEditMode ? 'Редактировать договор' : 'Новый договор' }}</h2>
      </template>

      <ContractForm
          v-if="isModalOpen"
          :initial-data="currentContract"
          @save="handleSave"
          @cancel="isModalOpen = false"
      />
    </Modal>

    <!-- Connection Modal -->
    <Modal :is-open="showConnectionModal" @close="showConnectionModal = false">
      <template #header>
        <h2>{{ isConnectionEditMode ? 'Редактировать подключение' : 'Новое подключение' }}</h2>
      </template>

      <ConnectionForm
        v-if="showConnectionModal"
        :initial-data="currentConnection"
        @save="handleConnectionSave"
        @cancel="showConnectionModal = false"
      />
    </Modal>

  </div>
</template>

<style scoped>
.btn-icon {
  width: 32px;
  height: 32px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
}

.block-btn {
  background: linear-gradient(135deg, var(--error-500) 0%, var(--error-600) 100%);
  color: white;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease-in-out;
  box-shadow: 0 2px 4px rgba(234, 67, 53, 0.2);
  border-radius: 50%;
}

.block-btn:hover {
  background: linear-gradient(135deg, var(--error-600) 0%, var(--error-700) 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(234, 67, 53, 0.3);
}

.unblock-btn {
  background: linear-gradient(135deg, var(--success-500) 0%, var(--success-600) 100%);
  color: white;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease-in-out;
  box-shadow: 0 2px 4px rgba(52, 168, 83, 0.2);
  border-radius: 50%;
}

.unblock-btn:hover {
  background: linear-gradient(135deg, var(--success-600) 0%, var(--success-700) 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(52, 168, 83, 0.3);
}

.stats-btn {
  background: linear-gradient(135deg, var(--primary-500) 0%, var(--primary-600) 100%);
  color: white;
  text-decoration: none;
  transition: all 0.2s ease-in-out;
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.2);
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.stats-btn:hover {
  background: linear-gradient(135deg, var(--primary-600) 0%, var(--primary-700) 100%);
  color: white;
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(59, 130, 246, 0.3);
}

.mr-2 {
  margin-right: 8px;
}

.connections-btn {
  background: linear-gradient(135deg, var(--warning-500) 0%, var(--warning-600) 100%);
  color: white;
  transition: all 0.2s ease-in-out;
  box-shadow: 0 2px 4px rgba(245, 158, 11, 0.2);
}

.connections-btn:hover {
  background: linear-gradient(135deg, var(--warning-600) 0%, var(--warning-700) 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(245, 158, 11, 0.3);
}

.connections-toggle {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: var(--radius-md);
  transition: background-color 0.2s ease;
}

.connections-toggle:hover {
  background: var(--gray-100);
}

.connections-toggle .count {
  font-weight: 500;
  color: var(--primary-600);
}

.toggle-icon {
  font-size: 1rem !important;
  color: var(--gray-500);
}

.connections-expanded {
  padding: 1rem;
  background: var(--gray-50);
  border-radius: var(--radius-md);
  margin: 0.5rem 0;
}

.loading {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  justify-content: center;
  padding: 1rem;
  color: var(--gray-500);
}

.no-connections {
  padding: 1rem;
  text-align: center;
  color: var(--gray-500);
}

.no-connections p {
  margin: 0 0 0.5rem 0;
}

.connections-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 0;
  margin-bottom: 0.75rem;
  border-bottom: 1px solid var(--gray-200);
}

.connections-list {
  max-height: none;
}

.connection-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem;
  border-bottom: 1px solid var(--gray-100);
}

.connection-item:last-child {
  border-bottom: none;
}

.connection-info {
  flex: 1;
}

.connection-main {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.25rem;
}

.connection-main .address {
  font-weight: 500;
  color: var(--gray-900);
}

.connection-main .ip {
  font-family: monospace;
  color: var(--primary-600);
  background: var(--primary-50);
  padding: 0.125rem 0.375rem;
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
}

.connection-details {
  display: flex;
  gap: 0.75rem;
  font-size: 0.75rem;
  color: var(--gray-600);
}

.connection-actions {
  display: flex;
  gap: 0.375rem;
  margin-left: 0.75rem;
}

.btn-xs {
  width: 24px;
  height: 24px;
  padding: 0;
  font-size: 0.625rem;
}

.btn-xs .icon {
  font-size: 0.625rem;
}

.edit-btn {
  background: linear-gradient(135deg, var(--primary-500) 0%, var(--primary-600) 100%);
  color: white;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease-in-out;
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.2);
  border-radius: 50%;
}

.edit-btn:hover {
  background: linear-gradient(135deg, var(--primary-600) 0%, var(--primary-700) 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(59, 130, 246, 0.3);
}
</style>