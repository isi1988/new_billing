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
const ipConflicts = ref([]);
const isCheckingIP = ref(false);
const validationErrors = ref([]);

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
    validationErrors.value.push("Ошибка загрузки списков для формы. Проверьте консоль.");
  } finally {
    isLoadingRelated.value = false;
  }
}

// Загружаем данные при монтировании компонента
onMounted(fetchRelatedData);

// Функция для проверки конфликтов IP адресов
async function checkIPConflicts() {
  const ipAddress = form.value.ip_address?.trim();
  
  // Очищаем конфликты если IP пустой
  if (!ipAddress) {
    ipConflicts.value = [];
    return;
  }

  // Простая валидация IP формата
  const ipRegex = /^(\d{1,3}\.){3}\d{1,3}$/;
  if (!ipRegex.test(ipAddress)) {
    ipConflicts.value = [];
    return;
  }

  isCheckingIP.value = true;
  try {
    // Получаем все подключения и проверяем IP конфликты
    const response = await apiClient.get('/connections');
    const allConnections = response.data || [];
    
    // Фильтруем подключения с таким же IP (исключая текущее при редактировании)
    const conflicts = allConnections.filter(conn => {
      return conn.ip_address === ipAddress && 
             conn.id !== form.value.id; // исключаем текущее подключение при редактировании
    });
    
    ipConflicts.value = conflicts;
  } catch (error) {
    console.error('Failed to check IP conflicts:', error);
    // При ошибке не показываем конфликты
    ipConflicts.value = [];
  } finally {
    isCheckingIP.value = false;
  }
}

// Отслеживаем изменения IP адреса с задержкой
let ipCheckTimeout;
watch(() => form.value.ip_address, () => {
  clearTimeout(ipCheckTimeout);
  ipCheckTimeout = setTimeout(checkIPConflicts, 500); // Задержка 500ms
});

function handleSubmit() {
  // Очищаем предыдущие ошибки
  validationErrors.value = [];

  // Валидация: проверяем, что все обязательные поля заполнены
  if (!form.value.equipment_id) {
    validationErrors.value.push('Пожалуйста, выберите оборудование');
  }
  if (!form.value.contract_id) {
    validationErrors.value.push('Пожалуйста, выберите договор');
  }
  if (!form.value.tariff_id) {
    validationErrors.value.push('Пожалуйста, выберите тариф');
  }

  // Если есть ошибки валидации, не продолжаем
  if (validationErrors.value.length > 0) {
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
    validationErrors.value.push('Некорректный ID оборудования');
  }
  if (isNaN(dataToSave.contract_id) || dataToSave.contract_id <= 0) {
    validationErrors.value.push('Некорректный ID договора');
  }
  if (isNaN(dataToSave.tariff_id) || dataToSave.tariff_id <= 0) {
    validationErrors.value.push('Некорректный ID тарифа');
  }

  // Если есть ошибки после преобразования, не продолжаем
  if (validationErrors.value.length > 0) {
    return;
  }

  emit('save', dataToSave);
}
</script>

<template>
  <form @submit.prevent="handleSubmit">
    <!-- Показываем ошибки валидации -->
    <div v-if="validationErrors.length > 0" class="validation-errors">
      <div class="error-header">
        <span class="material-icons icon-sm">error</span>
        Исправьте ошибки:
      </div>
      <ul>
        <li v-for="error in validationErrors" :key="error">{{ error }}</li>
      </ul>
    </div>

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
        
        <!-- IP conflict warning -->
        <div v-if="isCheckingIP" class="ip-status checking">
          ⏳ Проверка конфликтов IP...
        </div>
        <div v-else-if="ipConflicts.length > 0" class="ip-status conflict">
          <div class="conflict-header">
            ⚠️ IP адрес {{ form.ip_address }} уже используется:
          </div>
          <div class="conflict-list">
            <div v-for="conflict in ipConflicts" :key="conflict.id" class="conflict-item">
              • Подключение ID {{ conflict.id }} (Договор {{ conflict.contract_id }}, {{ conflict.address }})
            </div>
          </div>
        </div>
        <div v-else-if="form.ip_address && form.ip_address.match(/^(\d{1,3}\.){3}\d{1,3}$/)" class="ip-status success">
          ✅ IP адрес свободен
        </div>
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

/* IP Status styles */
.ip-status {
  margin-top: 8px;
  padding: 8px 12px;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 500;
}

.ip-status.checking {
  background-color: #fef3c7;
  color: #d97706;
  border: 1px solid #fde68a;
}

.ip-status.success {
  background-color: #f0fdf4;
  color: #15803d;
  border: 1px solid #bbf7d0;
}

.ip-status.conflict {
  background-color: #fef2f2;
  color: #dc2626;
  border: 1px solid #fecaca;
}

.conflict-header {
  font-weight: 600;
  margin-bottom: 6px;
}

.conflict-list {
  margin-left: 8px;
}

.conflict-item {
  margin-bottom: 4px;
  font-size: 13px;
}
</style>