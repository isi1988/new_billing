<script setup>
import { ref, computed, reactive } from 'vue';
import { useCrud } from '@/composables/useCrud';
import DataTable from '@/components/ui/DataTable.vue';
import Modal from '@/components/ui/Modal.vue';
import EquipmentForm from '@/components/forms/EquipmentForm.vue';
import SearchFilters from '@/components/ui/SearchFilters.vue';
import { useNotificationStore } from '@/stores/notification';

const notificationStore = useNotificationStore();

// Инициализируем CRUD-операции для эндпоинта '/api/equipment'
const {
  items: equipment,
  loading,
  createItem,
  updateItem,
  deleteItem
} = useCrud('equipment');

// Состояние для управления модальным окном
const isModalOpen = ref(false);
const currentEquipment = ref(null);
const isEditMode = ref(false);

// Search and filter state
const searchQuery = ref('');
const filterValues = reactive({});

// No filters for equipment - just search
const filters = [];

// Computed filtered equipment
const filteredEquipment = computed(() => {
  let filtered = equipment.value;

  // Apply search
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(item => {
      const model = (item.model || '').toLowerCase();
      const macAddress = (item.mac_address || '').toLowerCase();
      const description = (item.description || '').toLowerCase();
      return model.includes(query) || macAddress.includes(query) || description.includes(query);
    });
  }

  return filtered;
});

// Описание колонок для нашей таблицы
const columns = [
  { key: 'id', label: 'ID' },
  { key: 'model', label: 'Модель' },
  { key: 'mac_address', label: 'MAC-адрес' },
  { key: 'description', label: 'Описание' },
];

// Функция для открытия модального окна в режиме "Создание"
function openCreateModal() {
  isEditMode.value = false;
  // Задаем пустой объект с полями по умолчанию
  currentEquipment.value = {
    model: '',
    mac_address: '',
    description: ''
  };
  isModalOpen.value = true;
}

// Функция для открытия модального окна в режиме "Редактирование"
function openEditModal(item) {
  isEditMode.value = true;
  // Копируем данные из строки таблицы в currentEquipment
  currentEquipment.value = { ...item };
  isModalOpen.value = true;
}

// Обработчик события 'save' от формы
async function handleSave(equipmentData) {
  try {
    if (isEditMode.value) {
      await updateItem(equipmentData.id, equipmentData);
      notificationStore.addNotification({
        type: 'success',
        title: 'Оборудование обновлено',
        message: 'Данные оборудования успешно обновлены'
      });
    } else {
      await createItem(equipmentData);
      notificationStore.addNotification({
        type: 'success',
        title: 'Оборудование создано',
        message: 'Новое оборудование успешно добавлено'
      });
    }
    isModalOpen.value = false; // Закрываем модальное окно при успехе
    currentEquipment.value = null; // Очищаем форму
  } catch (error) {
    notificationStore.addNotification({
      type: 'error',
      title: 'Ошибка сохранения',
      message: 'Не удалось сохранить данные'
    });
  }
}

// Обработчик события 'delete' от таблицы
async function handleDelete(itemId) {
  try {
    await deleteItem(itemId);
    notificationStore.addNotification({
      type: 'success',
      title: 'Оборудование удалено',
      message: 'Оборудование успешно удалено'
    });
  } catch (error) {
    notificationStore.addNotification({
      type: 'error',
      title: 'Ошибка удаления',
      message: 'Не удалось удалить запись'
    });
  }
}

// Search and filter functions
function clearFilters() {
  searchQuery.value = '';
}
</script>

<template>
  <div class="page-container">
    <header class="page-header">
      <h1>Управление оборудованием</h1>
    </header>

    <SearchFilters
      :search-query="searchQuery"
      search-placeholder="Поиск по модели, MAC-адресу или описанию..."
      :filters="filters"
      :filter-values="filterValues"
      @search="searchQuery = $event"
      @filter="filterValues[$event.key] = $event.value"
      @clear="clearFilters"
    />

    <!-- Компонент таблицы данных -->
    <DataTable
        :items="filteredEquipment"
        :columns="columns"
        :loading="loading"
        @edit="openEditModal"
        @delete="handleDelete"
    />

    <!-- Кнопка для добавления новой записи -->
    <button class="fab" @click="openCreateModal">
      <span class="material-icons icon-lg">add</span>
    </button>

    <!-- Модальное окно для формы -->
    <Modal :is-open="isModalOpen" @close="isModalOpen = false">
      <template #header>
        <h2>{{ isEditMode ? 'Редактировать оборудование' : 'Новое оборудование' }}</h2>
      </template>

      <!-- v-if нужен, чтобы форма пересоздавалась каждый раз при открытии модалки -->
      <EquipmentForm
          v-if="isModalOpen"
          :initial-data="currentEquipment"
          @save="handleSave"
          @cancel="isModalOpen = false"
      />
    </Modal>
  </div>
</template>