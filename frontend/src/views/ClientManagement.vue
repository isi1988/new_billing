<script setup>
import { ref, computed, reactive } from 'vue';
import { useRouter } from 'vue-router';
import { useCrud } from '@/composables/useCrud';
import DataTable from '@/components/ui/DataTable.vue';
import Modal from '@/components/ui/Modal.vue';
import ClientForm from '@/components/forms/ClientForm.vue';
import StatusBadge from '@/components/ui/StatusBadge.vue';
import PhoneDisplay from '@/components/ui/PhoneDisplay.vue';
import SearchFilters from '@/components/ui/SearchFilters.vue';
import apiClient from '@/api/client';
import { formatDate } from '@/utils/dateUtils';

// Инициализируем роутер
const router = useRouter();

// Инициализируем CRUD для эндпоинта '/api/clients'
const {
  items: clients,
  loading,
  createItem,
  updateItem,
  deleteItem
} = useCrud('clients');

// Состояние для расширяемых договоров
const expandedClients = ref(new Set());
const clientContracts = ref({});
const loadingContracts = ref(new Set());

const isModalOpen = ref(false);
const currentClient = ref(null);
const isEditMode = ref(false);

// Search and filter state
const searchQuery = ref('');
const filterValues = reactive({
  client_type: '',
  is_blocked: ''
});

// Filter configuration
const filters = [
  {
    key: 'client_type',
    label: 'Тип клиента',
    type: 'select',
    options: [
      { value: 'individual', label: 'Физическое лицо' },
      { value: 'legal_entity', label: 'Юридическое лицо' }
    ]
  },
  {
    key: 'is_blocked',
    label: 'Статус',
    type: 'select',
    options: [
      { value: 'false', label: 'Активные' },
      { value: 'true', label: 'Заблокированные' }
    ]
  }
];

// Computed filtered clients
const filteredClients = computed(() => {
  let filtered = clients.value;

  // Apply search
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(client => {
      const name = getClientName(client).toLowerCase();
      const email = (client.email || '').toLowerCase();
      const phone = (client.phone || '').toLowerCase();
      return name.includes(query) || email.includes(query) || phone.includes(query);
    });
  }

  // Apply filters
  if (filterValues.client_type) {
    filtered = filtered.filter(client => client.client_type === filterValues.client_type);
  }
  if (filterValues.is_blocked !== '') {
    const isBlocked = filterValues.is_blocked === 'true';
    filtered = filtered.filter(client => client.is_blocked === isBlocked);
  }

  return filtered;
});

// Helper function to get client display name
function getClientName(client) {
  if (client.client_type === 'individual') {
    return `${client.last_name || ''} ${client.first_name || ''}`.trim();
  }
  return client.short_name || client.full_name || 'Юр. лицо';
}

// Колонки для таблицы
const columns = [
  { key: 'id', label: 'ID' },
  {
    key: 'name',
    label: 'Имя / Название',
    formatter: (client) => getClientName(client)
  },
  { key: 'client_type', label: 'Тип', component: 'StatusBadge' },
  { key: 'email', label: 'Email' },
  { key: 'phone', label: 'Телефон', component: 'PhoneDisplay' },
  { key: 'contracts_count', label: 'Договоров' },
  {
    key: 'is_blocked',
    label: 'Статус',
    component: 'StatusBadge'
  },
];

function openCreateModal() {
  isEditMode.value = false;
  // Данные по умолчанию для нового клиента
  currentClient.value = {
    client_type: 'individual',
    email: '',
    phone: '',
  };
  isModalOpen.value = true;
}

function openEditModal(item) {
  isEditMode.value = true;
  currentClient.value = { ...item };
  isModalOpen.value = true;
}

async function handleSave(clientData) {
  try {
    if (isEditMode.value) {
      await updateItem(clientData.id, clientData);
    } else {
      await createItem(clientData);
    }
    isModalOpen.value = false;
    currentClient.value = null; // Очищаем форму
  } catch (error) {
    alert('Не удалось сохранить клиента.');
  }
}

async function handleDelete(itemId) {
  if (confirm('Вы уверены, что хотите удалить клиента? Все его договоры и подключения также будут удалены!')) {
    try {
      await deleteItem(itemId);
    } catch (error) {
      alert('Не удалось удалить клиента.');
    }
  }
}

async function handleBlock(client) {
  const action = client.is_blocked ? 'разблокировать' : 'заблокировать';
  if (confirm(`Вы уверены, что хотите ${action} клиента?`)) {
    try {
      const endpoint = client.is_blocked ? 'unblock' : 'block';
      await apiClient.post(`/clients/${client.id}/${endpoint}`);
      
      // Обновляем статус клиента в локальном массиве
      const index = clients.value.findIndex(c => c.id === client.id);
      if (index !== -1) {
        clients.value[index].is_blocked = !client.is_blocked;
      }
    } catch (error) {
      alert(`Не удалось ${action} клиента.`);
    }
  }
}

// Search and filter functions
// Функция для переключения состояния расширения договоров
async function toggleContracts(client) {
  if (expandedClients.value.has(client.id)) {
    expandedClients.value.delete(client.id);
  } else {
    expandedClients.value.add(client.id);
    
    if (!clientContracts.value[client.id]) {
      loadingContracts.value.add(client.id);
      
      try {
        const response = await apiClient.get(`/clients/${client.id}/contracts`);
        clientContracts.value[client.id] = response.data || [];
      } catch (error) {
        console.error('Ошибка загрузки договоров:', error);
        alert('Не удалось загрузить договоры.');
        clientContracts.value[client.id] = [];
      } finally {
        loadingContracts.value.delete(client.id);
      }
    }
  }
}

// Функция блокировки/разблокировки договора
async function handleContractBlock(contract, clientId) {
  const action = contract.is_blocked ? 'разблокировать' : 'заблокировать';
  if (confirm(`Вы уверены, что хотите ${action} договор?`)) {
    try {
      const endpoint = contract.is_blocked ? 'unblock' : 'block';
      await apiClient.post(`/contracts/${contract.id}/${endpoint}`);
      
      // Обновляем статус договора в локальном массиве
      if (clientContracts.value[clientId]) {
        const index = clientContracts.value[clientId].findIndex(c => c.id === contract.id);
        if (index !== -1) {
          clientContracts.value[clientId][index].is_blocked = !contract.is_blocked;
        }
      }
    } catch (error) {
      alert(`Не удалось ${action} договор.`);
    }
  }
}

// Функции для управления договорами
function createContract(client) {
  router.push({
    path: '/contracts',
    query: {
      create: 'true',
      client_id: client.id,
      return_to: 'clients'
    }
  });
}

function editContract(contract) {
  router.push({
    path: '/contracts',
    query: {
      edit: contract.id,
      return_to: 'clients'
    }
  });
}

function clearFilters() {
  searchQuery.value = '';
  filterValues.client_type = '';
  filterValues.is_blocked = '';
}
</script>

<template>
  <div class="page-container">
    <header class="page-header">
      <h1>Управление клиентами</h1>
    </header>

    <SearchFilters
      :search-query="searchQuery"
      search-placeholder="Поиск по имени, email или телефону..."
      :filters="filters"
      :filter-values="filterValues"
      @search="searchQuery = $event"
      @filter="filterValues[$event.key] = $event.value"
      @clear="clearFilters"
    />

    <DataTable
        :items="filteredClients"
        :columns="columns"
        :loading="loading"
        @edit="openEditModal"
        @delete="handleDelete"
    >
      <template #cell-client_type="{ item }">
        <StatusBadge type="client_type" :value="item.client_type" size="small" />
      </template>
      
      <template #cell-phone="{ item }">
        <PhoneDisplay :phone="item.phone" format="russian" />
      </template>
      
      <template #cell-contracts_count="{ item }">
        <button 
          @click="toggleContracts(item)"
          class="contracts-toggle"
          :class="{ 'expanded': expandedClients.has(item.id) }"
        >
          <span class="count">{{ item.contracts_count || '0' }}</span>
          <span class="material-icons toggle-icon">{{ expandedClients.has(item.id) ? 'expand_less' : 'expand_more' }}</span>
        </button>
      </template>

      <template #cell-is_blocked="{ item }">
        <StatusBadge type="blocked_status" :value="item.is_blocked" size="small" />
      </template>
      
      <template #actions="{ item }">
        <button 
          @click="handleBlock(item)" 
          :class="['btn btn-icon btn-sm', item.is_blocked ? 'unblock-btn' : 'block-btn']"
          :title="item.is_blocked ? 'Разблокировать клиента' : 'Заблокировать клиента'"
        >
          <span class="material-icons icon-sm">{{ item.is_blocked ? 'lock_open' : 'lock' }}</span>
        </button>
      </template>

      <!-- Expandable rows for contracts -->
      <template v-for="item in filteredClients" :key="item.id" #[`expand-${item.id}`]>
        <div v-if="expandedClients.has(item.id)" class="contracts-expanded">
          <div v-if="loadingContracts.has(item.id)" class="loading">
            <span class="material-icons">hourglass_empty</span>
            Загрузка договоров...
          </div>
          <div v-else-if="clientContracts[item.id]?.length === 0" class="no-contracts">
            <p>Нет договоров</p>
            <button @click.stop="createContract(item)" class="btn btn-sm btn-primary">
              <span class="material-icons">add</span> Добавить
            </button>
          </div>
          <div v-else class="contracts-list">
            <div class="contracts-header">
              <strong>Договоры ({{ clientContracts[item.id]?.length || 0 }})</strong>
              <button @click.stop="createContract(item)" class="btn btn-xs btn-primary">
                <span class="material-icons">add</span> Добавить
              </button>
            </div>
            <div v-for="contract in clientContracts[item.id]" :key="contract.id" class="contract-item">
              <div class="contract-info">
                <div class="contract-main">
                  <span class="number">{{ contract.number }}</span>
                  <span class="sign-date">{{ formatDate(contract.sign_date) }}</span>
                  <StatusBadge type="blocked_status" :value="contract.is_blocked" size="small" />
                </div>
                <div class="contract-details">
                  <span class="detail">Подключений: {{ contract.connections_count || 0 }}</span>
                </div>
              </div>
              <div class="contract-actions">
                <button 
                  @click.stop="editContract(contract)"
                  class="btn btn-icon btn-xs edit-btn"
                  title="Редактировать"
                >
                  <span class="material-icons">edit</span>
                </button>
                <button 
                  @click.stop="handleContractBlock(contract, item.id)" 
                  :class="['btn btn-icon btn-xs', contract.is_blocked ? 'unblock-btn' : 'block-btn']"
                  :title="contract.is_blocked ? 'Разблокировать' : 'Заблокировать'"
                >
                  <span class="material-icons">{{ contract.is_blocked ? 'lock_open' : 'lock' }}</span>
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
        <h2>{{ isEditMode ? 'Редактировать клиента' : 'Новый клиент' }}</h2>
      </template>

      <ClientForm
          v-if="isModalOpen"
          :initial-data="currentClient"
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
  border: none;
  cursor: pointer;
  transition: all 0.2s ease-in-out;
  box-shadow: 0 2px 4px rgba(234, 67, 53, 0.2);
  border-radius: 50%;
  width: 24px;
  height: 24px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
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
  width: 24px;
  height: 24px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.unblock-btn:hover {
  background: linear-gradient(135deg, var(--success-600) 0%, var(--success-700) 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(52, 168, 83, 0.3);
}

.contracts-toggle {
  background: none;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--primary-600);
  font-weight: 500;
  padding: 4px 8px;
  border-radius: 4px;
  transition: all 0.2s ease-in-out;
}

.contracts-toggle:hover {
  background: var(--primary-50);
  color: var(--primary-700);
}

.contracts-toggle.expanded {
  color: var(--primary-700);
}

.toggle-icon {
  transition: transform 0.2s ease-in-out;
}

.contracts-toggle.expanded .toggle-icon {
  transform: rotate(180deg);
}

.contracts-expanded {
  padding: 16px;
  background: var(--gray-50);
  border-top: 1px solid var(--gray-200);
}

.loading {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--gray-600);
  padding: 20px;
  justify-content: center;
}

.no-contracts {
  text-align: center;
  padding: 20px;
  color: var(--gray-600);
}

.contracts-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.contracts-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.contract-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: white;
  border: 1px solid var(--gray-200);
  border-radius: 8px;
  transition: all 0.2s ease-in-out;
}

.contract-item:hover {
  border-color: var(--primary-300);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.contract-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.contract-main {
  display: flex;
  align-items: center;
  gap: 12px;
}

.contract-main .number {
  font-weight: 600;
  color: var(--gray-900);
}

.contract-main .sign-date {
  color: var(--gray-600);
  font-size: 14px;
}

.contract-details {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: var(--gray-600);
}

.contract-actions {
  display: flex;
  gap: 8px;
}

.edit-btn {
  background: linear-gradient(135deg, var(--primary-500) 0%, var(--primary-600) 100%);
  color: white;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease-in-out;
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.2);
  border-radius: 50%;
  width: 24px;
  height: 24px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.edit-btn:hover {
  background: linear-gradient(135deg, var(--primary-600) 0%, var(--primary-700) 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(59, 130, 246, 0.3);
}
</style>