<script setup>
defineProps({
  items: { type: Array, required: true },
  columns: { type: Array, required: true },
  loading: { type: Boolean, default: false },
});

const emit = defineEmits(['edit', 'delete']);
</script>

<template>
  <div class="data-table-container">
    <div v-if="loading" class="loading-spinner">Загрузка...</div>
    <table v-else class="data-table">
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
      <tr v-for="item in items" :key="item.id">
        <td v-for="col in columns" :key="col.key">
          {{ col.formatter ? col.formatter(item) : item[col.key] }}
        </td>
        <td>
          <button @click="emit('edit', item)" class="action-btn edit-btn">✎</button>
          <button @click="emit('delete', item.id)" class="action-btn delete-btn">🗑</button>
        </td>
      </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.data-table-container {
  background: var(--surface-color);
  border-radius: 8px;
  box-shadow: var(--shadow);
  overflow-x: auto; /* Позволяет таблице скроллиться по горизонтали на мал. экранах */
}
.data-table {
  width: 100%;
  border-collapse: collapse;
}
.data-table th, .data-table td {
  padding: 16px;
  text-align: left;
  border-bottom: 1px solid var(--border-color);
  white-space: nowrap; /* Предотвращает перенос строк в ячейках */
}
.data-table th {
  font-weight: 500;
  font-size: 14px;
  color: var(--text-color-light);
  text-transform: uppercase;
  background-color: #F7FAFC;
}
.data-table tbody tr:hover {
  background-color: #f9fafb;
}
.no-data {
  text-align: center;
  color: var(--text-color-light);
  padding: 32px;
  font-size: 16px;
}
.action-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 18px;
  padding: 4px;
  margin-right: 8px;
  transition: color 0.2s;
}
.edit-btn {
  color: var(--primary-color);
}
.edit-btn:hover {
  color: var(--primary-color-dark);
}
.delete-btn {
  color: var(--danger-color);
}
.delete-btn:hover {
  color: #c53030;
}
.loading-spinner {
  padding: 48px;
  text-align: center;
  font-size: 16px;
  color: var(--text-color-light);
}
</style>