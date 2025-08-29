<script setup>
import { ref, watch } from 'vue';

const props = defineProps({
  initialData: { type: Object, default: () => ({}) },
  isEditMode: { type: Boolean, default: false },
});
const emit = defineEmits(['save', 'cancel']);

const form = ref({});

watch(() => props.initialData, (newData) => {
  form.value = { ...newData };
}, { immediate: true, deep: true });

function handleSubmit() {
  emit('save', form.value);
}
</script>

<template>
  <form @submit.prevent="handleSubmit">
    <div class="form-group">
      <label for="title">Название задачи</label>
      <input id="title" type="text" v-model="form.title" required placeholder="Краткое описание задачи" />
    </div>
    <div class="form-group">
      <label for="description">Описание</label>
      <textarea id="description" v-model="form.description" required placeholder="Подробное описание доработки"></textarea>
    </div>
    <div class="form-actions">
      <button type="button" class="btn btn-secondary" @click="$emit('cancel')">Отмена</button>
      <button type="submit" class="btn btn-primary">Сохранить</button>
    </div>
  </form>
</template>