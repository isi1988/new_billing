<script setup>
import { ref, watch } from 'vue';

const props = defineProps({
  // Данные для предзаполнения формы в режиме редактирования
  initialData: { type: Object, default: () => ({}) },
});

const emit = defineEmits(['save', 'cancel']);

// Локальное состояние формы
const form = ref({});

// Эта функция следит за изменением props.initialData
// и обновляет локальное состояние формы.
// Это позволяет использовать одну и ту же форму для создания и редактирования.
watch(() => props.initialData, (newData) => {
  form.value = { ...newData };
}, {
  immediate: true, // Сработать сразу при создании компонента
  deep: true       // Глубокое отслеживание объекта
});

function handleSubmit() {
  emit('save', form.value);
}
</script>

<template>
  <form @submit.prevent="handleSubmit">
    <div class="form-group">
      <label for="model">Модель оборудования</label>
      <input id="model" type="text" v-model="form.model" required placeholder="Например, TP-Link Archer C80" />
    </div>

    <div class="form-group">
      <label for="mac-address">MAC-адрес</label>
      <input id="mac-address" type="text" v-model="form.mac_address" required placeholder="XX:XX:XX:XX:XX:XX" />
    </div>

    <div class="form-group">
      <label for="description">Описание</label>
      <textarea id="description" v-model="form.description" placeholder="Любая дополнительная информация..."></textarea>
    </div>

    <div class="form-actions">
      <button type="button" class="btn btn-secondary" @click="$emit('cancel')">Отмена</button>
      <button type="submit" class="btn btn-primary">Сохранить</button>
    </div>
  </form>
</template>

<style scoped>
textarea {
  min-height: 80px;
  resize: vertical;
}
</style>