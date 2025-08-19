<script setup>
import { ref } from 'vue';
import { useCrud } from '@/composables/useCrud';
import DataTable from '@/components/ui/DataTable.vue';
import Modal from '@/components/ui/Modal.vue';
import TariffForm from '@/components/forms/TariffForm.vue';

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

// Описание колонок для таблицы
const columns = [
  { key: 'id', label: 'ID' },
  { key: 'name', label: 'Название' },
  { key: 'payment_type', label: 'Тип оплаты' },
  { key: 'max_speed_in', label: 'Скорость вх. (Кбит/с)' },
  { key: 'is_archived', label: 'Архивный' },
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
    } else {
      await createItem(tariffData);
    }
    isModalOpen.value = false;
  } catch (error) {
    alert('Не удалось сохранить тариф.');
  }
}

// Обработка удаления тарифа
async function handleDelete(itemId) {
  if (confirm('Вы уверены, что хотите удалить этот тариф?')) {
    try {
      await deleteItem(itemId);
    } catch (error) {
      alert('Не удалось удалить тариф.');
    }
  }
}
</script>

<template>
  <div class="page-container">
    <header class="page-header">
      <h1>Управление тарифами</h1>
    </header>

    <DataTable
        :items="tariffs"
        :columns="columns"
        :loading="loading"
        @edit="openEditModal"
        @delete="handleDelete"
    />

    <button class="fab" @click="openCreateModal">+</button>

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
  </div>
</template>