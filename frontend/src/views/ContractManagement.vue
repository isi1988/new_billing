<script setup>
import { ref } from 'vue';
import { useCrud } from '@/composables/useCrud';
import DataTable from '@/components/ui/DataTable.vue';
import Modal from '@/components/ui/Modal.vue';
import ContractForm from '@/components/forms/ContractForm.vue';

// Инициализируем CRUD-операции для эндпоинта '/api/contracts'
const {
  items: contracts,
  loading,
  createItem,
  updateItem,
  deleteItem
} = useCrud('contracts');

// Состояние для управления модальным окном
const isModalOpen = ref(false);
const currentContract = ref(null);
const isEditMode = ref(false);

// Описание колонок для таблицы
const columns = [
  { key: 'id', label: 'ID' },
  { key: 'number', label: 'Номер' },
  { key: 'sign_date', label: 'Дата подписания', formatter: (date) => new Date(date).toLocaleDateString() },
  { key: 'client_id', label: 'ID Клиента' },
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
    } else {
      await createItem(contractData);
    }
    isModalOpen.value = false;
  } catch (error) {
    alert('Не удалось сохранить договор.');
  }
}

// Обработка удаления
async function handleDelete(itemId) {
  if (confirm('Вы уверены, что хотите удалить этот договор? Это также удалит все связанные с ним подключения.')) {
    try {
      await deleteItem(itemId);
    } catch (error) {
      alert('Не удалось удалить договор.');
    }
  }
}
</script>

<template>
  <div class="page-container">
    <header class="page-header">
      <h1>Управление договорами</h1>
    </header>

    <DataTable
        :items="contracts"
        :columns="columns"
        :loading="loading"
        @edit="openEditModal"
        @delete="handleDelete"
    />

    <button class="fab" @click="openCreateModal">+</button>

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
  </div>
</template>