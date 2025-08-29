<script setup>
import { ref, computed, reactive } from 'vue';
import { useCrud } from '@/composables/useCrud';
import DataTable from '@/components/ui/DataTable.vue';
import Modal from '@/components/ui/Modal.vue';
import TariffForm from '@/components/forms/TariffForm.vue';
import StatusBadge from '@/components/ui/StatusBadge.vue';
import SearchFilters from '@/components/ui/SearchFilters.vue';
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue';
import { useNotificationStore } from '@/stores/notification';

const notificationStore = useNotificationStore();

// Инициализируем CRUD-операции для эндпоинта '/api/tariffs'
const {
  items: tariffs,
  loading,
  createItem,
  updateItem,
  deleteItem
} = useCrud('tariffs');

// Состояние для управления модальным окном
const isModalOpen = ref(false);
const currentTariff = ref(null);
const isEditMode = ref(false);

// Confirm dialog state
const isDeleteDialogOpen = ref(false);
const tariffToDelete = ref(null);

// Search and filter state
const searchQuery = ref('');
const filterValues = reactive({
  payment_type: '',
  is_archived: '',
  is_for_individuals: ''
});

// Filter configuration
const filters = [
  {
    key: 'payment_type',
    label: 'Тип оплаты',
    type: 'select',
    options: [
      { value: 'prepaid', label: 'Предоплата' },
      { value: 'postpaid', label: 'Постоплата' }
    ]
  },
  {
    key: 'is_archived',
    label: 'Статус',
    type: 'select',
    options: [
      { value: 'false', label: 'Активные' },
      { value: 'true', label: 'Архивные' }
    ]
  },
  {
    key: 'is_for_individuals',
    label: 'Для кого',
    type: 'select',
    options: [
      { value: 'true', label: 'Физ. лица' },
      { value: 'false', label: 'Юр. лица' }
    ]
  }
];

// Computed filtered tariffs
const filteredTariffs = computed(() => {
  let filtered = tariffs.value;

  // Apply search
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(tariff => {
      const name = (tariff.name || '').toLowerCase();
      return name.includes(query);
    });
  }

  // Apply filters
  if (filterValues.payment_type) {
    filtered = filtered.filter(tariff => tariff.payment_type === filterValues.payment_type);
  }
  if (filterValues.is_archived !== '') {
    const isArchived = filterValues.is_archived === 'true';
    filtered = filtered.filter(tariff => tariff.is_archived === isArchived);
  }
  if (filterValues.is_for_individuals !== '') {
    const forIndividuals = filterValues.is_for_individuals === 'true';
    filtered = filtered.filter(tariff => tariff.is_for_individuals === forIndividuals);
  }

  return filtered;
});

// Описание колонок для таблицы
const columns = [
  { key: 'id', label: 'ID' },
  { key: 'name', label: 'Название' },
  { key: 'payment_type', label: 'Тип оплаты', component: 'StatusBadge' },
  { key: 'max_speed_in', label: 'Скорость вх. (Кбит/с)', formatter: (tariff) => `${(tariff.max_speed_in / 1000).toFixed(0)} Мбит/с` },
  { key: 'is_archived', label: 'Архивный', formatter: (tariff) => tariff.is_archived ? 'Архивный' : 'Активный' },
];

// Открытие модального окна для создания нового тарифа
function openCreateModal() {
  isEditMode.value = false;
  // Данные по умолчанию для нового тарифа
  currentTariff.value = {
    name: '',
    is_archived: false,
    payment_type: 'prepaid',
    is_for_individuals: true,
    max_speed_in: 100000, // 100 Мбит/с
    max_speed_out: 100000,
    max_traffic_in: 0,
    max_traffic_out: 0,
  };
  isModalOpen.value = true;
}

// Открытие модального окна для редактирования существующего тарифа
function openEditModal(item) {
  isEditMode.value = true;
  currentTariff.value = { ...item };
  isModalOpen.value = true;
}

// Обработка сохранения данных из формы
async function handleSave(tariffData) {
  try {
    if (isEditMode.value) {
      await updateItem(tariffData.id, tariffData);
      notificationStore.addNotification({
        type: 'success',
        title: 'Тариф обновлён',
        message: 'Данные тарифа успешно обновлены'
      });
    } else {
      await createItem(tariffData);
      notificationStore.addNotification({
        type: 'success',
        title: 'Тариф создан',
        message: 'Новый тариф успешно создан'
      });
    }
    isModalOpen.value = false;
    currentTariff.value = null; // Очищаем форму
  } catch (error) {
    notificationStore.addNotification({
      type: 'error',
      title: 'Ошибка сохранения',
      message: 'Не удалось сохранить тариф'
    });
  }
}

// Обработка удаления тарифа
function confirmDelete(itemId) {
  const tariff = tariffs.value.find(t => t.id === itemId);
  tariffToDelete.value = tariff;
  isDeleteDialogOpen.value = true;
}

async function handleDelete() {
  try {
    await deleteItem(tariffToDelete.value.id);
    notificationStore.addNotification({
      type: 'success',
      title: 'Тариф удалён',
      message: 'Тариф успешно удалён из системы'
    });
  } catch (error) {
    notificationStore.addNotification({
      type: 'error',
      title: 'Ошибка удаления',
      message: 'Не удалось удалить тариф'
    });
  } finally {
    isDeleteDialogOpen.value = false;
    tariffToDelete.value = null;
  }
}

function cancelDelete() {
  isDeleteDialogOpen.value = false;
  tariffToDelete.value = null;
}

// Search and filter functions
function clearFilters() {
  searchQuery.value = '';
  filterValues.payment_type = '';
  filterValues.is_archived = '';
  filterValues.is_for_individuals = '';
}
</script>

<template>
  <div class="page-container">
    <header class="page-header">
      <h1>Управление тарифами</h1>
    </header>

    <SearchFilters
      :search-query="searchQuery"
      search-placeholder="Поиск по названию тарифа..."
      :filters="filters"
      :filter-values="filterValues"
      @search="searchQuery = $event"
      @filter="filterValues[$event.key] = $event.value"
      @clear="clearFilters"
    />

    <DataTable
        :items="filteredTariffs"
        :columns="columns"
        :loading="loading"
        @edit="openEditModal"
        @delete="confirmDelete"
    >
      <template #cell-payment_type="{ item }">
        <StatusBadge type="payment_type" :value="item.payment_type" size="small" />
      </template>
    </DataTable>

    <button class="fab" @click="openCreateModal">
      <span class="material-icons icon-lg">add</span>
    </button>

    <Modal :is-open="isModalOpen" @close="isModalOpen = false">
      <template #header>
        <h2>{{ isEditMode ? 'Редактировать тариф' : 'Новый тариф' }}</h2>
      </template>

      <TariffForm
          v-if="isModalOpen"
          :initial-data="currentTariff"
          @save="handleSave"
          @cancel="isModalOpen = false"
      />
    </Modal>

    <ConfirmDialog
      :is-open="isDeleteDialogOpen"
      type="danger"
      title="Подтвердите удаление"
      :message="tariffToDelete ? `Вы действительно хотите удалить тариф '${tariffToDelete.name}'?` : ''"
      details="Это действие нельзя отменить. Все связанные подключения потеряют связь с этим тарифом."
      confirm-text="Удалить"
      cancel-text="Отмена"
      @confirm="handleDelete"
      @cancel="cancelDelete"
    />
  </div>
</template>