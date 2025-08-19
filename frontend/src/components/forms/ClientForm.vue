<script setup>
import { ref, watch } from 'vue';

const props = defineProps({
  initialData: { type: Object, default: () => ({}) },
});
const emit = defineEmits(['save', 'cancel']);

const form = ref({});

watch(() => props.initialData, (newData) => {
  // Форматируем даты для input type="date"
  const formattedData = { ...newData };
  if (formattedData.passport_issue_date) {
    formattedData.passport_issue_date = formattedData.passport_issue_date.split('T')[0];
  }
  if (formattedData.birth_date) {
    formattedData.birth_date = formattedData.birth_date.split('T')[0];
  }
  if (formattedData.ogrn_date) {
    formattedData.ogrn_date = formattedData.ogrn_date.split('T')[0];
  }
  form.value = formattedData;
}, { immediate: true, deep: true });

function handleSubmit() {
  emit('save', form.value);
}
</script>

<template>
  <form @submit.prevent="handleSubmit">
    <!-- Основные поля -->
    <div class="form-grid">
      <div class="form-group">
        <label for="client-type">Тип клиента</label>
        <select id="client-type" v-model="form.client_type" required>
          <option value="individual">Физическое лицо</option>
          <option value="legal_entity">Юридическое лицо</option>
        </select>
      </div>
      <div class="form-group">
        <label for="email">Email</label>
        <input id="email" type="email" v-model="form.email" required />
      </div>
      <div class="form-group">
        <label for="phone">Телефон</label>
        <input id="phone" type="tel" v-model="form.phone" required />
      </div>
    </div>

    <hr class="form-divider" />

    <!-- Поля для ФИЗИЧЕСКОГО ЛИЦА -->
    <div v-if="form.client_type === 'individual'" class="form-section">
      <h3>Данные физического лица</h3>
      <div class="form-grid">
        <div class="form-group"><label for="last_name">Фамилия</label><input id="last_name" type="text" v-model="form.last_name" /></div>
        <div class="form-group"><label for="first_name">Имя</label><input id="first_name" type="text" v-model="form.first_name" /></div>
        <div class="form-group"><label for="patronymic">Отчество</label><input id="patronymic" type="text" v-model="form.patronymic" /></div>
        <div class="form-group"><label for="birth_date">Дата рождения</label><input id="birth_date" type="date" v-model="form.birth_date" /></div>
        <div class="form-group span-2"><label for="registration_address">Адрес прописки</label><input id="registration_address" type="text" v-model="form.registration_address" /></div>
        <div class="form-group"><label for="passport_number">Паспорт (серия, номер)</label><input id="passport_number" type="text" v-model="form.passport_number" /></div>
        <div class="form-group"><label for="passport_issue_date">Дата выдачи</label><input id="passport_issue_date" type="date" v-model="form.passport_issue_date" /></div>
        <div class="form-group span-2"><label for="passport_issued_by">Кем выдан</label><input id="passport_issued_by" type="text" v-model="form.passport_issued_by" /></div>
      </div>
    </div>

    <!-- Поля для ЮРИДИЧЕСКОГО ЛИЦА -->
    <div v-if="form.client_type === 'legal_entity'" class="form-section">
      <h3>Данные юридического лица</h3>
      <div class="form-grid">
        <div class="form-group span-2"><label for="full_name">Полное наименование</label><input id="full_name" type="text" v-model="form.full_name" /></div>
        <div class="form-group"><label for="short_name">Краткое наименование</label><input id="short_name" type="text" v-model="form.short_name" /></div>
        <div class="form-group"><label for="inn">ИНН</label><input id="inn" type="text" v-model="form.inn" /></div>
        <div class="form-group"><label for="kpp">КПП</label><input id="kpp" type="text" v-model="form.kpp" /></div>
        <div class="form-group"><label for="ogrn">ОГРН</label><input id="ogrn" type="text" v-model="form.ogrn" /></div>
        <div class="form-group"><label for="ogrn_date">Дата ОГРН</label><input id="ogrn_date" type="date" v-model="form.ogrn_date" /></div>
        <div class="form-group span-2"><label for="legal_address">Юридический адрес</label><input id="legal_address" type="text" v-model="form.legal_address" /></div>
        <div class="form-group span-2"><label for="actual_address">Фактический адрес</label><input id="actual_address" type="text" v-model="form.actual_address" /></div>
        <div class="form-group span-2"><label for="bank_details">Банковские реквизиты</label><textarea id="bank_details" v-model="form.bank_details"></textarea></div>
        <div class="form-group"><label for="ceo">Генеральный директор</label><input id="ceo" type="text" v-model="form.ceo" /></div>
        <div class="form-group"><label for="accountant">Главный бухгалтер</label><input id="accountant" type="text" v-model="form.accountant" /></div>
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
.form-divider {
  border: none;
  border-top: 1px solid var(--border-color);
  margin: 24px 0;
}
.form-section h3 {
  font-size: 16px;
  font-weight: 500;
  margin-top: 0;
  margin-bottom: 16px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-color);
}
textarea {
  min-height: 80px;
  resize: vertical;
}
</style>