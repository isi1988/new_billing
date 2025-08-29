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
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue';
import apiClient from '@/api/client';
import { formatDate } from '@/utils/dateUtils';
import { useNotificationStore } from '@/stores/notification';

const notificationStore = useNotificationStore();

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

// Confirm dialog state
const isDeleteDialogOpen = ref(false);
const isBlockDialogOpen = ref(false);
const isContractBlockDialogOpen = ref(false);
const clientToDelete = ref(null);
const clientToBlock = ref(null);
const contractToBlock = ref(null);
const contractBlockClientId = ref(null);

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
      notificationStore.addNotification({
        type: 'success',
        title: 'Клиент обновлён',
        message: 'Данные клиента успешно обновлены'
      });
    } else {
      await createItem(clientData);
      notificationStore.addNotification({
        type: 'success',
        title: 'Клиент создан',
        message: 'Новый клиент успешно создан'
      });
    }
    isModalOpen.value = false;
    currentClient.value = null; // Очищаем форму
  } catch (error) {
    notificationStore.addNotification({
      type: 'error',
      title: 'Ошибка сохранения',
      message: 'Не удалось сохранить клиента'
    });
  }
}

function confirmDelete(itemId) {
  const client = clients.value.find(c => c.id === itemId);
  clientToDelete.value = client;
  isDeleteDialogOpen.value = true;
}

async function handleDelete() {
  try {
    await deleteItem(clientToDelete.value.id);
    notificationStore.addNotification({
      type: 'success',
      title: 'Клиент удалён',
      message: 'Клиент успешно удалён'
    });
  } catch (error) {
    notificationStore.addNotification({
      type: 'error',
      title: 'Ошибка удаления',
      message: 'Не удалось удалить клиента'
    });
  } finally {
    isDeleteDialogOpen.value = false;
    clientToDelete.value = null;
  }
}

function cancelDelete() {
  isDeleteDialogOpen.value = false;
  clientToDelete.value = null;
}

function confirmBlock(client) {
  clientToBlock.value = client;
  isBlockDialogOpen.value = true;
}

async function handleBlock() {
  const action = clientToBlock.value.is_blocked ? 'разблокировать' : 'заблокировать';
  try {
    const endpoint = clientToBlock.value.is_blocked ? 'unblock' : 'block';
    await apiClient.post(`/clients/${clientToBlock.value.id}/${endpoint}`);
    
    // Обновляем статус клиента в локальном массиве
    const index = clients.value.findIndex(c => c.id === clientToBlock.value.id);
    if (index !== -1) {
      clients.value[index].is_blocked = !clientToBlock.value.is_blocked;
    }
    
    notificationStore.addNotification({
      type: 'success',
      title: 'Операция выполнена',
      message: `Клиент успешно ${clientToBlock.value.is_blocked ? 'разблокирован' : 'заблокирован'}`
    });
  } catch (error) {
    notificationStore.addNotification({
      type: 'error',
      title: 'Ошибка операции',
      message: `Не удалось ${action} клиента`
    });
  } finally {
    isBlockDialogOpen.value = false;
    clientToBlock.value = null;
  }
}

function cancelBlock() {
  isBlockDialogOpen.value = false;
  clientToBlock.value = null;
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
        notificationStore.addNotification({
          type: 'error',
          title: 'Ошибка загрузки',
          message: 'Не удалось загрузить договоры'
        });
        clientContracts.value[client.id] = [];
      } finally {
        loadingContracts.value.delete(client.id);
      }
    }
  }
}

// Функция блокировки/разблокировки договора
function confirmContractBlock(contract, clientId) {
  contractToBlock.value = contract;
  contractBlockClientId.value = clientId;
  isContractBlockDialogOpen.value = true;
}

async function handleContractBlock() {
  const action = contractToBlock.value.is_blocked ? 'разблокировать' : 'заблокировать';
  try {
    const endpoint = contractToBlock.value.is_blocked ? 'unblock' : 'block';
    await apiClient.post(`/contracts/${contractToBlock.value.id}/${endpoint}`);
    
    // Обновляем статус договора в локальном массиве
    if (clientContracts.value[contractBlockClientId.value]) {
      const index = clientContracts.value[contractBlockClientId.value].findIndex(c => c.id === contractToBlock.value.id);
      if (index !== -1) {
        clientContracts.value[contractBlockClientId.value][index].is_blocked = !contractToBlock.value.is_blocked;
      }
    }
    
    notificationStore.addNotification({
      type: 'success',
      title: 'Операция выполнена',
      message: `Договор успешно ${contractToBlock.value.is_blocked ? 'разблокирован' : 'заблокирован'}`
    });
  } catch (error) {
    notificationStore.addNotification({
      type: 'error',
      title: 'Ошибка операции',
      message: `Не удалось ${action} договор`
    });
  } finally {
    isContractBlockDialogOpen.value = false;
    contractToBlock.value = null;
    contractBlockClientId.value = null;
  }
}

function cancelContractBlock() {
  isContractBlockDialogOpen.value = false;
  contractToBlock.value = null;
  contractBlockClientId.value = null;
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
        @delete="confirmDelete"
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
          <span class="material-icons icon-sm toggle-icon">{{ expandedClients.has(item.id) ? 'expand_less' : 'expand_more' }}</span>
        </button>
      </template>

      <template #cell-is_blocked="{ item }">
        <StatusBadge type="blocked_status" :value="item.is_blocked" size="small" />
      </template>
      
      <template #actions="{ item }">
        <button 
          @click="confirmBlock(item)" 
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
            <span class="material-icons icon-sm">hourglass_empty</span>
            Загрузка договоров...
          </div>
          <div v-else-if="clientContracts[item.id]?.length === 0" class="no-contracts">
            <p>Нет договоров</p>
            <button @click.stop="createContract(item)" class="btn btn-sm btn-primary">
              <span class="material-icons icon-sm">add</span> Добавить
            </button>
          </div>
          <div v-else class="contracts-list">
            <div class="contracts-header">
              <strong>Договоры ({{ clientContracts[item.id]?.length || 0 }})</strong>
              <button @click.stop="createContract(item)" class="btn btn-sm btn-primary">
                <span class="material-icons icon-sm">add</span> Добавить
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
                  class="btn btn-icon btn-sm edit-btn"
                  title="Редактировать"
                >
                  <span class="material-icons icon-sm">edit</span>
                </button>
                <button 
                  @click.stop="confirmContractBlock(contract, item.id)" 
                  :class="['btn btn-icon btn-sm', contract.is_blocked ? 'unblock-btn' : 'block-btn']"
                  :title="contract.is_blocked ? 'Разблокировать' : 'Заблокировать'"
                >
                  <span class="material-icons icon-sm">{{ contract.is_blocked ? 'lock_open' : 'lock' }}</span>
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

    <!-- Delete Confirmation Dialog -->
    <ConfirmDialog
      :is-open="isDeleteDialogOpen"
      type="danger"
      title="Подтвердите удаление"
      :message="clientToDelete ? `Вы действительно хотите удалить клиента '${getClientName(clientToDelete)}'?` : ''"
      details="Это действие нельзя отменить. Все договоры и подключения клиента будут удалены навсегда."
      confirm-text="Удалить"
      cancel-text="Отмена"
      @confirm="handleDelete"
      @cancel="cancelDelete"
    />

    <!-- Block/Unblock Confirmation Dialog -->
    <ConfirmDialog
      :is-open="isBlockDialogOpen"
      :type="clientToBlock?.is_blocked ? 'info' : 'warning'"
      :title="clientToBlock?.is_blocked ? 'Разблокировать клиента' : 'Заблокировать клиента'"
      :message="clientToBlock ? `Вы действительно хотите ${clientToBlock.is_blocked ? 'разблокировать' : 'заблокировать'} клиента '${getClientName(clientToBlock)}'?` : ''"
      :details="clientToBlock?.is_blocked ? 'Клиент сможет снова пользоваться услугами.' : 'Клиент не сможет пользоваться услугами до разблокировки.'"
      :confirm-text="clientToBlock?.is_blocked ? 'Разблокировать' : 'Заблокировать'"
      cancel-text="Отмена"
      @confirm="handleBlock"
      @cancel="cancelBlock"
    />

    <!-- Contract Block/Unblock Confirmation Dialog -->
    <ConfirmDialog
      :is-open="isContractBlockDialogOpen"
      :type="contractToBlock?.is_blocked ? 'info' : 'warning'"
      :title="contractToBlock?.is_blocked ? 'Разблокировать договор' : 'Заблокировать договор'"
      :message="contractToBlock ? `Вы действительно хотите ${contractToBlock.is_blocked ? 'разблокировать' : 'заблокировать'} договор '${contractToBlock.number}'?` : ''"
      :details="contractToBlock?.is_blocked ? 'Все подключения по договору будут активированы.' : 'Все подключения по договору будут заблокированы.'"
      :confirm-text="contractToBlock?.is_blocked ? 'Разблокировать' : 'Заблокировать'"
      cancel-text="Отмена"
      @confirm="handleContractBlock"
      @cancel="cancelContractBlock"
    />
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