<script setup>
import { ref } from 'vue';
import { useCrud } from '@/composables/useCrud';
import DataTable from '@/components/ui/DataTable.vue';
import Modal from '@/components/ui/Modal.vue';
import UserForm from '@/components/forms/UserForm.vue';

const { items: users, loading, createItem, updateItem, deleteItem } = useCrud('users');

const isModalOpen = ref(false);
const currentUser = ref(null);
const isEditMode = ref(false);

const columns = [
  { key: 'id', label: 'ID' },
  { key: 'username', label: 'Имя пользователя' },
  { key: 'role', label: 'Роль' },
];

function openCreateModal() {
  isEditMode.value = false;
  currentUser.value = { username: '', password: '', role: 'manager' }; // Данные по умолчанию
  isModalOpen.value = true;
}

function openEditModal(user) {
  isEditMode.value = true;
  currentUser.value = { ...user };
  isModalOpen.value = true;
}

async function handleSave(userData) {
  try {
    if (isEditMode.value) {
      await updateItem(userData.id, userData);
    } else {
      await createItem(userData);
    }
    isModalOpen.value = false;
  } catch (error) {
    alert('Не удалось сохранить пользователя.');
  }
}

async function handleDelete(userId) {
  if (confirm('Вы уверены, что хотите удалить этого пользователя?')) {
    try {
      await deleteItem(userId);
    } catch (error) {
      alert('Не удалось удалить пользователя.');
    }
  }
}
</script>

<template>
  <div class="page-container">
    <header class="page-header">
      <h1>Управление пользователями</h1>
    </header>

    <DataTable
        :items="users"
        :columns="columns"
        :loading="loading"
        @edit="openEditModal"
        @delete="handleDelete"
    />

    <button class="fab" @click="openCreateModal">+</button>

    <Modal :is-open="isModalOpen" @close="isModalOpen = false">
      <template #header>
        <h2>{{ isEditMode ? 'Редактировать пользователя' : 'Создать пользователя' }}</h2>
      </template>
      <UserForm
          v-if="currentUser"
          :initial-data="currentUser"
          :is-edit-mode="isEditMode"
          @save="handleSave"
          @cancel="isModalOpen = false"
      />
    </Modal>
  </div>
</template>