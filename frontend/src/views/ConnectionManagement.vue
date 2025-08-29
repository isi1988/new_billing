<script setup>
import { ref, computed, reactive, onMounted } from 'vue';
import { useCrud } from '@/composables/useCrud';
import DataTable from '@/components/ui/DataTable.vue';
import Modal from '@/components/ui/Modal.vue';
import ConnectionForm from '@/components/forms/ConnectionForm.vue';
import SearchFilters from '@/components/ui/SearchFilters.vue';
import StatusBadge from '@/components/ui/StatusBadge.vue';
import apiClient from '@/api/client';

// Инициализируем CRUD-операции для эндпоинта '/api/connections'
const {
  items: connections,
  loading,
  createItem,
  updateItem,
  deleteItem
} = useCrud('connections');

// Состояние для управления модальным окном
const isModalOpen = ref(false);
const currentConnection = ref(null);
const isEditMode = ref(false);

// Search and filter state
const searchQuery = ref('');
const filterValues = reactive({
  connection_type: '',
  contract_number: '',
  tariff_name: ''
});

// Data for searchable dropdowns
const contracts = ref([]);
const tariffs = ref([]);
const loadingFilters = ref(false);

// Filter configuration
const filters = computed(() => [
  {
    key: 'connection_type',
    label: 'Тип подключения',
    type: 'select',
    options: [
      { value: 'FTTB', label: 'FTTB' },
      { value: 'FTTH', label: 'FTTH' },
      { value: 'DSL', label: 'DSL' },
      { value: 'Cable', label: 'Cable' },
      { value: 'Wireless', label: 'Wireless' }
    ]
  },
  {
    key: 'contract_number',
    label: 'Номер договора',
    type: 'searchable-select',
    options: contracts.value.map(contract => ({
      value: contract.number,
      label: `№${contract.number}`,
      searchText: `${contract.number} ${contract.client_name || ''}`
    })),
    placeholder: 'Поиск по номеру договора...',
    loading: loadingFilters.value
  },
  {
    key: 'tariff_name',
    label: 'Тариф',
    type: 'searchable-select', 
    options: tariffs.value.map(tariff => ({
      value: tariff.name,
      label: tariff.name,
      searchText: tariff.name
    })),
    placeholder: 'Поиск по названию тарифа...',
    loading: loadingFilters.value
  }
]);

// Computed filtered connections
const filteredConnections = computed(() => {
  let filtered = connections.value;

  // Apply search
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(connection => {
      const address = (connection.address || '').toLowerCase();
      const ipAddress = (connection.ip_address || '').toLowerCase();
      return address.includes(query) || ipAddress.includes(query);
    });
  }

  // Apply filters
  if (filterValues.connection_type) {
    filtered = filtered.filter(connection => connection.connection_type === filterValues.connection_type);
  }
  if (filterValues.contract_number) {
    filtered = filtered.filter(connection => connection.contract_number === filterValues.contract_number);
  }
  if (filterValues.tariff_name) {
    filtered = filtered.filter(connection => connection.tariff_name === filterValues.tariff_name);
  }

  return filtered;
});

// Описание колонок для таблицы
const columns = [
  { key: 'id', label: 'ID' },
  { key: 'address', label: 'Адрес' },
  { key: 'ip_address', label: 'IP-адрес' },
  { key: 'connection_type', label: 'Тип подключения' },
  { 
    key: 'contract_number', 
    label: 'Договор',
    formatter: (connection) => connection.contract_number || `ID: ${connection.contract_id}`
  },
  { 
    key: 'tariff_name', 
    label: 'Тариф',
    formatter: (connection) => connection.tariff_name || `ID: ${connection.tariff_id}`
  },
  { 
    key: 'equipment_model', 
    label: 'Оборудование',
    formatter: (connection) => connection.equipment_model || `ID: ${connection.equipment_id}`
  },
  {
    key: 'is_blocked',
    label: 'Статус',
    component: 'StatusBadge'
  },
];

// Открытие модального окна для создания нового подключения
function openCreateModal() {
  isEditMode.value = false;
  currentConnection.value = {
    address: '',
    ip_address: '',
    mask: 24,
    connection_type: 'FTTB',
    equipment_id: null,
    contract_id: null,
    tariff_id: null,
  };
  isModalOpen.value = true;
}

// Открытие модального окна для редактирования
function openEditModal(item) {
  isEditMode.value = true;
  currentConnection.value = { ...item };
  isModalOpen.value = true;
}

// Обработка сохранения данных из формы
async function handleSave(connectionData) {
  try {
    if (isEditMode.value) {
      await updateItem(connectionData.id, connectionData);
    } else {
      await createItem(connectionData);
    }
    isModalOpen.value = false;
    currentConnection.value = null; // Очищаем форму
  } catch (error) {
    alert('Не удалось сохранить подключение.');
  }
}

// Обработка удаления
async function handleDelete(itemId) {
  if (confirm('Вы уверены, что хотите удалить это подключение?')) {
    try {
      await deleteItem(itemId);
    } catch (error) {
      alert('Не удалось удалить подключение.');
    }
  }
}

// Функция блокировки/разблокировки подключения
async function handleBlock(connection) {
  const action = connection.is_blocked ? 'разблокировать' : 'заблокировать';
  if (confirm(`Вы уверены, что хотите ${action} подключение?`)) {
    try {
      const endpoint = connection.is_blocked ? 'unblock' : 'block';
      await apiClient.post(`/connections/${connection.id}/${endpoint}`);
      
      // Обновляем статус подключения в локальном массиве
      const index = connections.value.findIndex(c => c.id === connection.id);
      if (index !== -1) {
        connections.value[index].is_blocked = !connection.is_blocked;
      }
    } catch (error) {
      alert(`Не удалось ${action} подключение.`);
    }
  }
}

// Function to load contracts and tariffs for filters
async function loadFilterData() {
  loadingFilters.value = true;
  try {
    const [contractsRes, tariffsRes] = await Promise.all([
      apiClient.get('/contracts'),
      apiClient.get('/tariffs')
    ]);
    contracts.value = contractsRes.data || [];
    tariffs.value = tariffsRes.data || [];
  } catch (error) {
    console.error('Failed to load filter data:', error);
  } finally {
    loadingFilters.value = false;
  }
}

// Load filter data on component mount
onMounted(() => {
  loadFilterData();
});

// Search and filter functions
function clearFilters() {
  searchQuery.value = '';
  filterValues.connection_type = '';
  filterValues.contract_number = '';
  filterValues.tariff_name = '';
}
</script>

<template>
  <div class="page-container">
    <header class="page-header">
      <h1>Управление подключениями</h1>
    </header>

    <SearchFilters
      :search-query="searchQuery"
      search-placeholder="Поиск по адресу или IP адресу..."
      :filters="filters"
      :filter-values="filterValues"
      @search="searchQuery = $event"
      @filter="filterValues[$event.key] = $event.value"
      @clear="clearFilters"
    />

    <DataTable
        :items="filteredConnections"
        :columns="columns"
        :loading="loading"
        @edit="openEditModal"
        @delete="handleDelete"
    >
      <template #cell-is_blocked="{ item }">
        <StatusBadge type="blocked_status" :value="item.is_blocked" size="small" />
      </template>
      
      <template #actions="{ item }">
        <router-link 
          :to="`/connections/${item.id}/stats`"
          class="btn btn-icon btn-sm stats-btn mr-2"
          title="Статистика по подключению"
        >
          <span class="material-icons icon-sm">analytics</span>
        </router-link>
        <button 
          @click="handleBlock(item)" 
          :class="['btn btn-icon btn-sm', item.is_blocked ? 'unblock-btn' : 'block-btn']"
          :title="item.is_blocked ? 'Разблокировать подключение' : 'Заблокировать подключение'"
        >
          <span class="material-icons icon-sm">{{ item.is_blocked ? 'lock_open' : 'lock' }}</span>
        </button>
      </template>
    </DataTable>

    <button class="fab" @click="openCreateModal">
      <span class="material-icons icon-lg">add</span>
    </button>

    <Modal :is-open="isModalOpen" @close="isModalOpen = false">
      <template #header>
        <h2>{{ isEditMode ? 'Редактировать подключение' : 'Новое подключение' }}</h2>
      </template>

      <ConnectionForm
          v-if="isModalOpen"
          :initial-data="currentConnection"
          @save="handleSave"
          @cancel="isModalOpen = false"
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
  transition: all 0.2s ease-in-out;
  box-shadow: 0 2px 4px rgba(234, 67, 53, 0.2);
}

.block-btn:hover {
  background: linear-gradient(135deg, var(--error-600) 0%, var(--error-700) 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(234, 67, 53, 0.3);
}

.unblock-btn {
  background: linear-gradient(135deg, var(--success-500) 0%, var(--success-600) 100%);
  color: white;
  transition: all 0.2s ease-in-out;
  box-shadow: 0 2px 4px rgba(52, 168, 83, 0.2);
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
</style>