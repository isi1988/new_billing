<script setup>
import { ref, watch } from 'vue';

const props = defineProps({
  initialData: { type: Object, default: () => ({}) },
  isEditMode: { type: Boolean, default: false },
});
const emit = defineEmits(['save', 'cancel']);

const form = ref({});

// Следим за изменениями initialData и обновляем форму
watch(() => props.initialData, (newData) => {
  form.value = { ...newData };
  // Не сбрасываем пароль при редактировании, если он не меняется
  if (props.isEditMode) {
    form.value.password = '';
  }
}, { immediate: true, deep: true });

function handleSubmit() {
  const dataToSave = { ...form.value };
  // Не отправляем пустой пароль при редактировании
  if (props.isEditMode && !dataToSave.password) {
    delete dataToSave.password; // Бэкенд не должен обновлять пароль, если поле пустое
  }
  emit('save', dataToSave);
}
</script>

<template>
  <form @submit.prevent="handleSubmit">
    <div class="form-group">
      <label for="username">Имя пользователя</label>
      <input id="username" type="text" v-model="form.username" required />
    </div>
    <div class="form-group" v-if="!isEditMode">
      <label for="password">Пароль</label>
      <input id="password" type="password" v-model="form.password" required />
    </div>
    <div class="form-group">
      <label for="role">Роль</label>
      <select id="role" v-model="form.role" required>
        <option value="manager">Менеджер</option>
        <option value="admin">Администратор</option>
      </select>
    </div>
    <p v-if="isEditMode" class="form-note">Оставьте поле пароля пустым, если не хотите его изменять.</p>
    <div class="form-actions">
      <button type="button" class="btn btn-secondary" @click="$emit('cancel')">Отмена</button>
      <button type="submit" class="btn btn-primary">Сохранить</button>
    </div>
  </form>
</template>

<style scoped>
.form-note { font-size: 12px; color: var(--text-color-light); }
</style>