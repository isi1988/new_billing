<script setup>
import { ref, watch } from 'vue';

const props = defineProps({
  initialData: { type: Object, default: () => ({}) },
});
const emit = defineEmits(['save', 'cancel']);

const form = ref({});

watch(() => props.initialData, (newData) => {
  form.value = { ...newData };
}, { immediate: true, deep: true });

function handleSubmit() {
  // Преобразуем числовые поля из строк в числа перед отправкой
  const dataToSave = {
    ...form.value,
    max_speed_in: parseInt(form.value.max_speed_in, 10) || 0,
    max_speed_out: parseInt(form.value.max_speed_out, 10) || 0,
    max_traffic_in: parseInt(form.value.max_traffic_in, 10) || 0,
    max_traffic_out: parseInt(form.value.max_traffic_out, 10) || 0,
  };
  emit('save', dataToSave);
}
</script>

<template>
  <form @submit.prevent="handleSubmit">
    <div class="form-grid">
      <div class="form-group span-2">
        <label for="name">Название тарифа</label>
        <input id="name" type="text" v-model="form.name" required placeholder="Например, Домашний Интернет 100" />
      </div>

      <div class="form-group">
        <label for="payment-type">Тип оплаты</label>
        <select id="payment-type" v-model="form.payment_type" required>
          <option value="prepaid">Предоплатный</option>
          <option value="postpaid">Постоплатный</option>
        </select>
      </div>

      <div class="form-group checkbox-group">
        <input id="is-for-individuals" type="checkbox" v-model="form.is_for_individuals" />
        <label for="is-for-individuals">Для физ. лиц</label>
      </div>

      <div class="form-group">
        <label for="max-speed-in">Макс. скорость (вх.), Кбит/с</label>
        <input id="max-speed-in" type="number" v-model="form.max_speed_in" required />
      </div>

      <div class="form-group">
        <label for="max-speed-out">Макс. скорость (исх.), Кбит/с</label>
        <input id="max-speed-out" type="number" v-model="form.max_speed_out" required />
      </div>

      <div class="form-group">
        <label for="max-traffic-in">Входящий трафик, Байты</label>
        <input id="max-traffic-in" type="number" v-model="form.max_traffic_in" required placeholder="0 - безлимит" />
      </div>

      <div class="form-group">
        <label for="max-traffic-out">Исходящий трафик, Байты</label>
        <input id="max-traffic-out" type="number" v-model="form.max_traffic_out" required placeholder="0 - безлимит" />
      </div>

      <div class="form-group checkbox-group span-2">
        <input id="is-archived" type="checkbox" v-model="form.is_archived" />
        <label for="is-archived">Архивный (недоступен для новых подключений)</label>
      </div>
    </div>

    <div class="form-actions">
      <button type="button" class="btn btn-secondary" @click="$emit('cancel')">Отмена</button>
      <button type="submit" class="btn btn-primary">Сохранить</button>
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
.checkbox-group {
  display: flex;
  align-items: center;
  padding-top: 24px; /* Для выравнивания с инпутами */
}
.checkbox-group input {
  width: auto;
  margin-right: 8px;
}
</style>