<script setup>
defineProps({
  items: { type: Array, required: true },
  columns: { type: Array, required: true },
  loading: { type: Boolean, default: false },
});

const emit = defineEmits(['edit', 'delete']);
</script>

<template>
  <div class="card">
    <div v-if="loading" class="loading-container">
      <div class="spinner"></div>
    </div>
    <div v-else class="table-container">
      <table class="table">
        <thead>
        <tr>
          <th v-for="col in columns" :key="col.key">{{ col.label }}</th>
          <th>Действия</th>
        </tr>
        </thead>
        <tbody>
        <tr v-if="items.length === 0">
          <td :colspan="columns.length + 1" class="no-data">Нет данных для отображения</td>
        </tr>
        <template v-for="item in items" :key="item.id">
          <tr>
            <td v-for="col in columns" :key="col.key">
              <slot :name="`cell-${col.key}`" :item="item" :value="item[col.key]">
                {{ col.formatter ? col.formatter(item) : item[col.key] }}
              </slot>
            </td>
            <td class="actions-cell">
              <div class="actions-group">
                <button @click="emit('edit', item)" class="btn btn-icon btn-sm edit-btn" title="Редактировать">
                  <span class="material-icons icon-sm">edit</span>
                </button>
                <button @click="emit('delete', item.id)" class="btn btn-icon btn-sm delete-btn" title="Удалить">
                  <span class="material-icons icon-sm">delete</span>
                </button>
                <!-- Custom actions slot -->
                <slot name="actions" :item="item"></slot>
              </div>
            </td>
          </tr>
          <!-- Expandable row -->
          <tr v-if="$slots[`expand-${item.id}`]" class="expandable-row">
            <td :colspan="columns.length + 1" class="expand-cell">
              <slot :name="`expand-${item.id}`" :item="item"></slot>
            </td>
          </tr>
        </template>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 3rem;
}

.table-container {
  overflow-x: auto;
}

.no-data {
  text-align: center;
  color: var(--gray-500);
  padding: 2rem;
  font-size: 1rem;
}

.actions-cell {
  white-space: nowrap;
  width: 1%;
}

.actions-group {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.btn-icon {
  width: 32px;
  height: 32px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
}

.edit-btn {
  background: linear-gradient(135deg, var(--primary-500) 0%, var(--primary-600) 100%);
  color: white;
  transition: all 0.2s ease-in-out;
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.2);
}

.edit-btn:hover {
  background: linear-gradient(135deg, var(--primary-600) 0%, var(--primary-700) 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(59, 130, 246, 0.3);
}

.delete-btn {
  background: linear-gradient(135deg, var(--error-500) 0%, var(--error-600) 100%);
  color: white;
  transition: all 0.2s ease-in-out;
  box-shadow: 0 2px 4px rgba(234, 67, 53, 0.2);
}

.delete-btn:hover {
  background: linear-gradient(135deg, var(--error-600) 0%, var(--error-700) 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(234, 67, 53, 0.3);
}

.expandable-row {
  background: var(--gray-50);
}

.expand-cell {
  padding: 0 !important;
  border-top: none;
}
</style>