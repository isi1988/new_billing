<template>
  <div class="card mb-6">
    <div class="card-body">
      <h2 class="card-title">Поиск и фильтры</h2>
      
      <!-- Search Input -->
      <div class="search-section">
        <div class="form-group">
          <label class="form-label">Поиск</label>
          <input 
            :value="searchQuery"
            @input="updateSearch"
            type="text"
            class="form-control"
            :placeholder="searchPlaceholder"
          />
        </div>
      </div>

      <!-- Dynamic Filters -->
      <div v-if="filters.length > 0" class="filters-section">
        <div class="filter-grid">
          <div 
            v-for="filter in filters" 
            :key="filter.key"
            class="form-group"
          >
            <label class="form-label">{{ filter.label }}</label>
            
            <!-- Select Filter -->
            <select 
              v-if="filter.type === 'select'"
              :value="filterValues[filter.key]"
              @change="updateFilter(filter.key, $event.target.value)"
              class="form-control"
            >
              <option value="">{{ filter.allText || 'Все' }}</option>
              <option 
                v-for="option in filter.options" 
                :key="option.value"
                :value="option.value"
              >
                {{ option.label }}
              </option>
            </select>

            <!-- Date Filter -->
            <input 
              v-else-if="filter.type === 'date'"
              :value="filterValues[filter.key]"
              @change="updateFilter(filter.key, $event.target.value)"
              type="date"
              class="form-control"
            />

            <!-- Number Filter -->
            <input 
              v-else-if="filter.type === 'number'"
              :value="filterValues[filter.key]"
              @change="updateFilter(filter.key, $event.target.value)"
              type="number"
              class="form-control"
              :placeholder="filter.placeholder"
            />

            <!-- Text Filter -->
            <input 
              v-else
              :value="filterValues[filter.key]"
              @change="updateFilter(filter.key, $event.target.value)"
              type="text"
              class="form-control"
              :placeholder="filter.placeholder"
            />
          </div>
        </div>
      </div>

      <!-- Action Buttons -->
      <div class="action-section">
        <button 
          @click="applyFilters"
          class="btn btn-md btn-primary"
        >
          <span class="material-icons icon-sm">search</span>
          Применить
        </button>
        <button 
          @click="clearFilters"
          class="btn btn-md btn-secondary"
        >
          <span class="material-icons icon-sm">clear</span>
          Очистить
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, watch } from 'vue';

const props = defineProps({
  searchQuery: { type: String, default: '' },
  searchPlaceholder: { type: String, default: 'Поиск...' },
  filters: { type: Array, default: () => [] },
  filterValues: { type: Object, default: () => ({}) }
});

const emit = defineEmits(['search', 'filter', 'apply', 'clear']);

function updateSearch(event) {
  emit('search', event.target.value);
}

function updateFilter(key, value) {
  emit('filter', { key, value });
}

function applyFilters() {
  emit('apply');
}

function clearFilters() {
  emit('clear');
}
</script>

<style scoped>
.search-section {
  margin-bottom: 1.5rem;
}

.filters-section {
  margin-bottom: 1.5rem;
}

.filter-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
}

.action-section {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

@media (max-width: 768px) {
  .filter-grid {
    grid-template-columns: 1fr;
  }
  
  .action-section {
    flex-direction: column;
  }
}
</style>