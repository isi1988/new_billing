<script setup>
import { ref, computed, reactive } from 'vue';
import { useCrud } from '@/composables/useCrud';
import { useNotificationStore } from '@/stores/notification';
import DataTable from '@/components/ui/DataTable.vue';
import Modal from '@/components/ui/Modal.vue';
import UserForm from '@/components/forms/UserForm.vue';
import SearchFilters from '@/components/ui/SearchFilters.vue';
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue';

const { items: users, loading, createItem, updateItem, deleteItem } = useCrud('users');
const notificationStore = useNotificationStore();

const isModalOpen = ref(false);
const currentUser = ref(null);
const isEditMode = ref(false);

// Confirm dialog state
const isConfirmDialogOpen = ref(false);
const userToDelete = ref(null);

// Search and filter state
const searchQuery = ref('');
const filterValues = reactive({
  role: ''
});

// Filter configuration
const filters = [
  {
    key: 'role',
    label: 'Роль',
    type: 'select',
    options: [
      { value: 'admin', label: 'Администратор' },
      { value: 'manager', label: 'Менеджер' }
    ]
  }
];

// Computed filtered users
const filteredUsers = computed(() => {
  let filtered = users.value;

  // Apply search
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(user => {
      const username = (user.username || '').toLowerCase();
      return username.includes(query);
    });
  }

  // Apply filters
  if (filterValues.role) {
    filtered = filtered.filter(user => user.role === filterValues.role);
  }

  return filtered;
});

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
      notificationStore.addNotification({
        type: 'success',
        title: 'Пользователь обновлён',
        message: 'Данные пользователя успешно обновлены'
      });
    } else {
      await createItem(userData);
      notificationStore.addNotification({
        type: 'success',
        title: 'Пользователь создан',
        message: 'Новый пользователь успешно создан'
      });
    }
    isModalOpen.value = false;
    currentUser.value = null; // Очищаем форму
  } catch (error) {
    notificationStore.addNotification({
      type: 'error',
      title: 'Ошибка сохранения',
      message: 'Не удалось сохранить пользователя'
    });
  }
}

function confirmDelete(userId) {
  const user = users.value.find(u => u.id === userId);
  userToDelete.value = user;
  isConfirmDialogOpen.value = true;
}

async function handleDelete() {
  try {
    await deleteItem(userToDelete.value.id);
    notificationStore.addNotification({
      type: 'success',
      title: 'Пользователь удалён',
      message: 'Пользователь успешно удалён из системы'
    });
  } catch (error) {
    notificationStore.addNotification({
      type: 'error',
      title: 'Ошибка удаления',
      message: 'Не удалось удалить пользователя'
    });
  } finally {
    isConfirmDialogOpen.value = false;
    userToDelete.value = null;
  }
}

function cancelDelete() {
  isConfirmDialogOpen.value = false;
  userToDelete.value = null;
}

// Search and filter functions
function clearFilters() {
  searchQuery.value = '';
  filterValues.role = '';
}
</script>

<template>
  <div class="page-container">
    <header class="page-header">
      <h1>Управление пользователями</h1>
    </header>

    <SearchFilters
      :search-query="searchQuery"
      search-placeholder="Поиск по имени пользователя..."
      :filters="filters"
      :filter-values="filterValues"
      @search="searchQuery = $event"
      @filter="filterValues[$event.key] = $event.value"
      @clear="clearFilters"
    />

    <DataTable
        :items="filteredUsers"
        :columns="columns"
        :loading="loading"
        @edit="openEditModal"
        @delete="confirmDelete"
    />

    <button class="fab" @click="openCreateModal">
      <span class="material-icons icon-lg">add</span>
    </button>

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

    <ConfirmDialog
      :is-open="isConfirmDialogOpen"
      type="danger"
      title="Подтвердите удаление"
      :message="userToDelete ? `Вы действительно хотите удалить пользователя '${userToDelete.username}'?` : ''"
      details="Это действие нельзя отменить. Все данные пользователя будут удалены навсегда."
      confirm-text="Удалить"
      cancel-text="Отмена"
      @confirm="handleDelete"
      @cancel="cancelDelete"
    />
  </div>
</template>