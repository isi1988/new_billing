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
const clientSearchQuery = ref('');
const showClientDropdown = ref(false);
const filteredClients = ref([]);

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
    filteredClients.value = clients.value;
    
    // Если есть выбранный клиент, устанавливаем его имя в поле поиска
    if (form.value.client_id) {
      const selectedClient = clients.value.find(c => c.id === form.value.client_id);
      if (selectedClient) {
        clientSearchQuery.value = getClientDisplayName(selectedClient);
      }
    }
  } catch (error) {
    console.error("Не удалось загрузить список клиентов:", error);
    alert("Ошибка загрузки списка клиентов.");
  } finally {
    isLoadingClients.value = false;
  }
}

// Функция для получения отображаемого имени клиента
function getClientDisplayName(client) {
  if (!client) return '';
  
  if (client.client_type === 'individual') {
    const lastName = client.last_name || '';
    const firstName = client.first_name || '';
    const name = `${lastName} ${firstName}`.trim();
    return name || client.email || `Клиент ID: ${client.id}`;
  } else {
    return client.short_name || client.full_name || client.email || `Клиент ID: ${client.id}`;
  }
}

// Функция поиска клиентов
function searchClients() {
  if (!clientSearchQuery.value.trim()) {
    filteredClients.value = clients.value;
    return;
  }
  
  const query = clientSearchQuery.value.toLowerCase();
  filteredClients.value = clients.value.filter(client => {
    const displayName = getClientDisplayName(client).toLowerCase();
    const email = (client.email || '').toLowerCase();
    const id = client.id.toString();
    return displayName.includes(query) || email.includes(query) || id.includes(query);
  });
}

// Функция выбора клиента
function selectClient(client) {
  form.value.client_id = client.id;
  clientSearchQuery.value = getClientDisplayName(client);
  showClientDropdown.value = false;
}

// Функция скрытия dropdown с задержкой
function hideClientDropdown() {
  setTimeout(() => {
    showClientDropdown.value = false;
  }, 200);
}

// Загружаем клиентов при создании компонента
onMounted(fetchClients);

function handleSubmit() {
  // Преобразуем ID клиента в число, дату оставляем в формате YYYY-MM-DD
  const dataToSave = {
    ...form.value,
    client_id: parseInt(form.value.client_id, 10),
    // Оставляем дату в формате YYYY-MM-DD для нашего CustomDate
    sign_date: form.value.sign_date || null,
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
        <div class="client-search-container">
          <input 
            v-model="clientSearchQuery"
            type="text"
            class="form-control"
            placeholder="Поиск клиента по имени, email или ID..."
            @input="searchClients"
            @focus="showClientDropdown = true"
            @blur="hideClientDropdown"
            required
          />
          <div v-if="showClientDropdown && filteredClients.length > 0" class="client-dropdown">
            <div 
              v-for="client in filteredClients" 
              :key="client.id"
              class="client-option"
              @mousedown="selectClient(client)"
            >
              <div class="client-name">
                {{ getClientDisplayName(client) }}
              </div>
              <div class="client-meta">
                <span class="client-email">{{ client.email || 'Нет email' }}</span>
                <span class="client-id">ID: {{ client.id }}</span>
              </div>
            </div>
          </div>
          <div v-else-if="showClientDropdown && clientSearchQuery && filteredClients.length === 0" class="client-dropdown">
            <div class="no-results">Клиенты не найдены</div>
          </div>
        </div>
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

.client-search-container {
  position: relative;
}

.client-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: white;
  border: 1px solid var(--gray-300);
  border-radius: var(--radius-lg);
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  max-height: 300px;
  overflow-y: auto;
  z-index: 9999;
  margin-top: 2px;
}

.client-option {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--gray-100);
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.client-option:hover {
  background-color: var(--primary-50);
}

.client-option:last-child {
  border-bottom: none;
}

.client-name {
  font-weight: 500;
  color: var(--gray-900);
  margin-bottom: 0.25rem;
}

.client-meta {
  display: flex;
  justify-content: space-between;
  font-size: 0.875rem;
}

.client-email {
  color: var(--gray-600);
}

.client-id {
  color: var(--gray-500);
  font-weight: 500;
}

.no-results {
  padding: 1rem;
  text-align: center;
  color: var(--gray-500);
}
</style>