<script setup>
import { ref, watch } from 'vue';

const props = defineProps({
  initialData: { type: Object, default: () => ({}) },
});
const emit = defineEmits(['save', 'cancel']);

const form = ref({});

function formatPhoneInput(event) {
  const input = event.target;
  let value = input.value.replace(/\D/g, ''); // Remove all non-digits
  
  if (value.startsWith('8') && value.length > 1) {
    value = '7' + value.slice(1);
  }
  
  if (value.startsWith('7') && value.length <= 11) {
    const match = value.match(/^7(\d{0,3})(\d{0,3})(\d{0,2})(\d{0,2})$/);
    if (match) {
      value = '+7';
      if (match[1]) value += ` (${match[1]}`;
      if (match[2]) value += `) ${match[2]}`;
      if (match[3]) value += `-${match[3]}`;
      if (match[4]) value += `-${match[4]}`;
    }
  }
  
  input.value = value;
  form.value.phone = value;
}

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
        <input 
          id="phone" 
          type="tel" 
          v-model="form.phone" 
          @input="formatPhoneInput"
          placeholder="+7 (999) 999-99-99"
          required 
        />
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
        
        <!-- Банковские реквизиты -->
        <div class="form-group span-2">
          <h4 class="form-subsection">Банковские реквизиты</h4>
        </div>
        <div class="form-group span-2"><label for="bank_name">Наименование банка</label><input id="bank_name" type="text" v-model="form.bank_name" placeholder="Например: ПАО Сбербанк" /></div>
        <div class="form-group"><label for="bank_account">Расчетный счет</label><input id="bank_account" type="text" v-model="form.bank_account" placeholder="40702810..." maxlength="20" /></div>
        <div class="form-group"><label for="bank_bik">БИК банка</label><input id="bank_bik" type="text" v-model="form.bank_bik" placeholder="044525225" maxlength="9" /></div>
        <div class="form-group span-2"><label for="bank_correspondent">Корреспондентский счет</label><input id="bank_correspondent" type="text" v-model="form.bank_correspondent" placeholder="30101810..." maxlength="20" /></div>
        
        <div class="form-group"><label for="ceo">Генеральный директор</label><input id="ceo" type="text" v-model="form.ceo" /></div>
        <div class="form-group"><label for="accountant">Главный бухгалтер</label><input id="accountant" type="text" v-model="form.accountant" /></div>
      </div>
    </div>

    <div class="form-actions">
      <button type="button" class="btn btn-md btn-secondary" @click="$emit('cancel')">
        <span class="icon icon-sm">❌</span>
        Отмена
      </button>
      <button type="submit" class="btn btn-md btn-primary">
        <span class="icon icon-sm">✓</span>
        Сохранить
      </button>
    </div>
  </form>
</template>

<style scoped>
.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
}

.span-2 {
  grid-column: span 2;
}

.form-divider {
  border: none;
  border-top: 1px solid var(--gray-200);
  margin: 1.5rem 0;
}

.form-section h3 {
  font-size: 1.125rem;
  font-weight: 600;
  margin-top: 0;
  margin-bottom: 1rem;
  padding-bottom: 0.5rem;
  border-bottom: 2px solid var(--primary-200);
  color: var(--gray-900);
}

textarea.form-control {
  min-height: 5rem;
  resize: vertical;
}

.form-subsection {
  font-size: 0.875rem;
  font-weight: 500;
  margin: 0;
  color: var(--gray-700);
  padding-bottom: 0.25rem;
  border-bottom: 1px dotted var(--gray-300);
}

@media (max-width: 768px) {
  .form-grid {
    grid-template-columns: 1fr;
  }
  
  .span-2 {
    grid-column: span 1;
  }
}
</style>