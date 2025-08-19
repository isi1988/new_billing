<script setup>
import { ref } from 'vue';
import { useCrud } from '@/composables/useCrud';
import DataTable from '@/components/ui/DataTable.vue';
import Modal from '@/components/ui/Modal.vue';
import ConnectionForm from '@/components/forms/ConnectionForm.vue';

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

// Описание колонок для таблицы.
// Примечание: мы показываем ID, так как для показа имен нужна более сложная логика.
const columns = [
  { key: 'id', label: 'ID' },
  { key: 'address', label: 'Адрес' },
  { key: 'ip_address', label: 'IP Адрес' },
  { key: 'tariff_id', label: 'ID Тарифа' },
  { key: 'contract_id', label: 'ID Договора' },
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
</script>

<template>
  <div class="page-container">
    <header class="page-header">
      <h1>Управление подключениями</h1>
    </header>

    <DataTable
        :items="connections"
        :columns="columns"
        :loading="loading"
        @edit="openEditModal"
        @delete="handleDelete"
    />

    <button class="fab" @click="openCreateModal">+</button>

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