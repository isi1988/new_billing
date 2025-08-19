<script setup>
import { ref } from 'vue';
import { useCrud } from '@/composables/useCrud'; // Наша универсальная CRUD-логика
import DataTable from '@/components/ui/DataTable.vue'; // Переиспользуемая таблица
import Modal from '@/components/ui/Modal.vue'; // Переиспользуемое модальное окно
import EquipmentForm from '@/components/forms/EquipmentForm.vue'; // Форма, которую мы создали выше

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
const currentEquipment = ref(null); // Здесь будет храниться объект для редактирования/создания
const isEditMode = ref(false);

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
    } else {
      await createItem(equipmentData);
    }
    isModalOpen.value = false; // Закрываем модальное окно при успехе
  } catch (error) {
    alert('Не удалось сохранить данные. Проверьте консоль.');
  }
}

// Обработчик события 'delete' от таблицы
async function handleDelete(itemId) {
  if (confirm('Вы уверены, что хотите удалить эту запись?')) {
    try {
      await deleteItem(itemId);
    } catch (error) {
      alert('Не удалось удалить запись. Проверьте консоль.');
    }
  }
}
</script>

<template>
  <div class="page-container">
    <header class="page-header">
      <h1>Управление оборудованием</h1>
    </header>

    <!-- Компонент таблицы данных -->
    <DataTable
        :items="equipment"
        :columns="columns"
        :loading="loading"
        @edit="openEditModal"
        @delete="handleDelete"
    />

    <!-- Кнопка для добавления новой записи -->
    <button class="fab" @click="openCreateModal">+</button>

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