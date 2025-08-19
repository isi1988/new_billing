<script setup>
import { ref, watch, onMounted } from 'vue';
import apiClient from '@/api/client';

const props = defineProps({
  initialData: { type: Object, default: () => ({}) },
});
const emit = defineEmits(['save', 'cancel']);

// --- Локальное состояние ---
const form = ref({});
const clients = ref([]); // Список клиентов для выпадающего меню
const isLoadingClients = ref(false);

// --- Логика ---

// Обновляем форму при изменении initialData
watch(() => props.initialData, (newData) => {
  // Форматируем дату для input type="date", который ожидает 'YYYY-MM-DD'
  const formattedData = { ...newData };
  if (formattedData.sign_date) {
    formattedData.sign_date = formattedData.sign_date.split('T')[0];
  }
  form.value = formattedData;
}, { immediate: true, deep: true });

// Функция для загрузки списка клиентов
async function fetchClients() {
  isLoadingClients.value = true;
  try {
    const response = await apiClient.get('/clients');
    clients.value = response.data || [];
  } catch (error) {
    console.error("Не удалось загрузить список клиентов:", error);
    alert("Ошибка загрузки списка клиентов.");
  } finally {
    isLoadingClients.value = false;
  }
}

// Загружаем клиентов при создании компонента
onMounted(fetchClients);

function handleSubmit() {
  // Преобразуем ID клиента в число
  const dataToSave = {
    ...form.value,
    client_id: parseInt(form.value.client_id, 10),
  };
  emit('save', dataToSave);
}
</script>

<template>
  <form @submit.prevent="handleSubmit">
    <div v-if="isLoadingClients" class="loading-related">Загрузка списка клиентов...</div>
    <div v-else class="form-grid">
      <div class="form-group span-2">
        <label for="client">Клиент</label>
        <select id="client" v-model="form.client_id" required>
          <option :value="null" disabled>-- Выберите клиента --</option>
          <option v-for="client in clients" :key="client.id" :value="client.id">
            <!-- Показываем разную информацию для физ. и юр. лиц -->
            {{ client.short_name || `${client.last_name} ${client.first_name}` }} (ID: {{ client.id }})
          </option>
        </select>
      </div>

      <div class="form-group">
        <label for="number">Номер договора</label>
        <input id="number" type="text" v-model="form.number" required placeholder="Например, 12345/2025" />
      </div>

      <div class="form-group">
        <label for="sign-date">Дата подписания</label>
        <input id="sign-date" type="date" v-model="form.sign_date" required />
      </div>
    </div>

    <div class="form-actions">
      <button type="button" class="btn btn-secondary" @click="$emit('cancel')">Отмена</button>
      <button type="submit" class="btn btn-primary" :disabled="isLoadingClients">Сохранить</button>
    </div>
  </form>
</template>

<style scoped>
.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
.span-2 {
  grid-column: span 2;
}
.loading-related {
  padding: 32px;
  text-align: center;
  color: var(--text-color-light);
}
</style>