<script setup>
import { ref } from 'vue';
import { useCrud } from '@/composables/useCrud';
import DataTable from '@/components/ui/DataTable.vue';
import Modal from '@/components/ui/Modal.vue';
import ClientForm from '@/components/forms/ClientForm.vue';

// Инициализируем CRUD для эндпоинта '/api/clients'
const {
  items: clients,
  loading,
  createItem,
  updateItem,
  deleteItem
} = useCrud('clients');

const isModalOpen = ref(false);
const currentClient = ref(null);
const isEditMode = ref(false);

// Колонки для таблицы. Добавляем "умный" форматер для имени.
const columns = [
  { key: 'id', label: 'ID' },
  {
    key: 'name',
    label: 'Имя / Название',
    // Форматер - функция для кастомного отображения данных в ячейке
    formatter: (client) => {
      if (client.client_type === 'individual') {
        return `${client.last_name || ''} ${client.first_name || ''}`.trim();
      }
      return client.short_name || client.full_name || 'Юр. лицо';
    }
  },
  { key: 'client_type', label: 'Тип' },
  { key: 'email', label: 'Email' },
  { key: 'phone', label: 'Телефон' },
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
</script>

<template>
  <div class="page-container">
    <header class="page-header">
      <h1>Управление клиентами</h1>
    </header>

    <DataTable
        :items="clients"
        :columns="columns"
        :loading="loading"
        @edit="openEditModal"
        @delete="handleDelete"
    />

    <button class="fab" @click="openCreateModal">+</button>

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