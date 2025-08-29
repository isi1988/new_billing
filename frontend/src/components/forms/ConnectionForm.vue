<script setup>
import { ref, watch, onMounted } from 'vue';
import apiClient from '@/api/client';

const props = defineProps({
  initialData: { type: Object, default: () => ({}) },
});
const emit = defineEmits(['save', 'cancel']);

// --- Локальное состояние ---
const form = ref({});
const relatedData = ref({
  equipment: [],
  contracts: [],
  tariffs: [],
});
const isLoadingRelated = ref(false);

// --- Логика ---

// Обновляем форму при изменении initialData
watch(() => props.initialData, (newData) => {
  form.value = { ...newData };
}, { immediate: true, deep: true });

// Функция для загрузки списков для выпадающих меню
async function fetchRelatedData() {
  isLoadingRelated.value = true;
  try {
    const [equipRes, contractsRes, tariffsRes] = await Promise.all([
      apiClient.get('/equipment'),
      apiClient.get('/contracts'),
      apiClient.get('/tariffs')
    ]);
    relatedData.value.equipment = equipRes.data || [];
    relatedData.value.contracts = contractsRes.data || [];
    relatedData.value.tariffs = tariffsRes.data || [];
  } catch (error) {
    console.error("Не удалось загрузить связанные данные:", error);
    alert("Ошибка загрузки списков для формы. Проверьте консоль.");
  } finally {
    isLoadingRelated.value = false;
  }
}

// Загружаем данные при монтировании компонента
onMounted(fetchRelatedData);

function handleSubmit() {
  // Валидация: проверяем, что все обязательные поля заполнены
  if (!form.value.equipment_id) {
    alert('Пожалуйста, выберите оборудование');
    return;
  }
  if (!form.value.contract_id) {
    alert('Пожалуйста, выберите договор');
    return;
  }
  if (!form.value.tariff_id) {
    alert('Пожалуйста, выберите тариф');
    return;
  }

  // Преобразуем числовые поля
  const dataToSave = {
    ...form.value,
    mask: parseInt(form.value.mask, 10) || 0,
    equipment_id: parseInt(form.value.equipment_id, 10),
    contract_id: parseInt(form.value.contract_id, 10),
    tariff_id: parseInt(form.value.tariff_id, 10),
  };
  
  // Проверяем, что ID действительно числа и больше 0
  if (isNaN(dataToSave.equipment_id) || dataToSave.equipment_id <= 0) {
    alert('Некорректный ID оборудования');
    return;
  }
  if (isNaN(dataToSave.contract_id) || dataToSave.contract_id <= 0) {
    alert('Некорректный ID договора');
    return;
  }
  if (isNaN(dataToSave.tariff_id) || dataToSave.tariff_id <= 0) {
    alert('Некорректный ID тарифа');
    return;
  }

  emit('save', dataToSave);
}
</script>

<template>
  <form @submit.prevent="handleSubmit">
    <div v-if="isLoadingRelated" class="loading-related">Загрузка списков...</div>
    <div v-else class="form-grid">
      <div class="form-group span-2">
        <label for="address">Адрес подключения</label>
        <input id="address" type="text" v-model="form.address" required placeholder="Город, улица, дом, квартира" />
      </div>

      <div class="form-group">
        <label for="contract">Договор</label>
        <select id="contract" v-model="form.contract_id" required>
          <option :value="null" disabled>Выберите договор</option>
          <option v-for="c in relatedData.contracts" :key="c.id" :value="c.id">
            №{{ c.number }} (Клиент ID: {{ c.client_id }})
          </option>
        </select>
      </div>

      <div class="form-group">
        <label for="tariff">Тариф</label>
        <select id="tariff" v-model="form.tariff_id" required>
          <option :value="null" disabled>Выберите тариф</option>
          <option v-for="t in relatedData.tariffs" :key="t.id" :value="t.id">
            {{ t.name }}
          </option>
        </select>
      </div>

      <div class="form-group">
        <label for="equipment">Оборудование</label>
        <select id="equipment" v-model="form.equipment_id" required :disabled="relatedData.equipment.length === 0">
          <option :value="null" disabled>
            {{ relatedData.equipment.length === 0 ? 'Нет доступного оборудования' : 'Выберите оборудование' }}
          </option>
          <option v-for="e in relatedData.equipment" :key="e.id" :value="e.id">
            {{ e.model }} ({{ e.mac_address }})
          </option>
        </select>
        <div v-if="!isLoadingRelated && relatedData.equipment.length === 0" class="error-message">
          ⚠️ В базе данных нет доступного оборудования. Обратитесь к администратору.
        </div>
      </div>

      <div class="form-group">
        <label for="connection-type">Тип подключения</label>
        <input id="connection-type" type="text" v-model="form.connection_type" placeholder="Например, FTTB" />
      </div>

      <div class="form-group">
        <label for="ip-address">IP-адрес</label>
        <input id="ip-address" type="text" v-model="form.ip_address" required placeholder="xxx.xxx.xxx.xxx" />
      </div>

      <div class="form-group">
        <label for="mask">Маска подсети</label>
        <input id="mask" type="number" v-model="form.mask" required min="0" max="32" />
      </div>
    </div>

    <div class="form-actions">
      <button type="button" class="btn btn-secondary" @click="$emit('cancel')">Отмена</button>
      <button 
        type="submit" 
        class="btn btn-primary" 
        :disabled="isLoadingRelated || relatedData.equipment.length === 0 || relatedData.contracts.length === 0 || relatedData.tariffs.length === 0"
      >
        Сохранить
      </button>
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
.error-message {
  margin-top: 4px;
  padding: 8px;
  background-color: #fef2f2;
  color: #dc2626;
  border: 1px solid #fecaca;
  border-radius: 4px;
  font-size: 14px;
}
</style>