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

            <!-- Searchable Select Filter -->
            <div v-else-if="filter.type === 'searchable-select'" class="searchable-select">
              <input 
                :value="getSelectedLabel(filter)"
                @input="onSearchableInput(filter.key, $event.target.value)"
                @focus="showDropdown(filter.key)"
                @blur="hideDropdown(filter.key)"
                type="text"
                class="form-control"
                :placeholder="filter.placeholder || 'Выберите...' "
                :disabled="filter.loading"
                autocomplete="off"
              />
              <div 
                v-if="dropdownStates[filter.key] && getFilteredOptions(filter).length > 0" 
                class="dropdown-menu"
              >
                <div 
                  v-for="option in getFilteredOptions(filter)" 
                  :key="option.value"
                  class="dropdown-item"
                  @mousedown="selectOption(filter.key, option.value, option.label)"
                >
                  {{ option.label }}
                </div>
              </div>
              <span v-if="filter.loading" class="loading-indicator">
                <span class="material-icons">hourglass_empty</span>
              </span>
            </div>

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
import { ref, reactive, watch, computed } from 'vue';

const props = defineProps({
  searchQuery: { type: String, default: '' },
  searchPlaceholder: { type: String, default: 'Поиск...' },
  filters: { type: Array, default: () => [] },
  filterValues: { type: Object, default: () => ({}) }
});

const emit = defineEmits(['search', 'filter', 'apply', 'clear']);

// State for searchable selects
const searchQueries = ref({});
const dropdownStates = ref({});
const selectedLabels = ref({});

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
  searchQueries.value = {};
  selectedLabels.value = {};
  dropdownStates.value = {};
  emit('clear');
}

// Searchable select functions
function onSearchableInput(key, value) {
  searchQueries.value[key] = value;
  if (!value.trim()) {
    updateFilter(key, '');
    selectedLabels.value[key] = '';
  }
}

function showDropdown(key) {
  dropdownStates.value[key] = true;
}

function hideDropdown(key) {
  setTimeout(() => {
    dropdownStates.value[key] = false;
  }, 150);
}

function selectOption(key, value, label) {
  updateFilter(key, value);
  selectedLabels.value[key] = label;
  searchQueries.value[key] = '';
  dropdownStates.value[key] = false;
}

function getSelectedLabel(filter) {
  const key = filter.key;
  if (selectedLabels.value[key]) {
    return selectedLabels.value[key];
  }
  if (props.filterValues[key]) {
    // Find the label for the current value
    const option = filter.options.find(opt => opt.value === props.filterValues[key]);
    return option ? option.label : '';
  }
  return searchQueries.value[key] || '';
}

function getFilteredOptions(filter) {
  if (!filter.options) return [];
  
  const query = searchQueries.value[filter.key]?.toLowerCase() || '';
  if (!query) return filter.options.slice(0, 10); // Limit to first 10 options
  
  return filter.options.filter(option => {
    const searchText = option.searchText || option.label;
    return searchText.toLowerCase().includes(query);
  }).slice(0, 10);
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

.searchable-select {
  position: relative;
}

.dropdown-menu {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: white;
  border: 1px solid var(--gray-300);
  border-radius: 4px;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
  max-height: 200px;
  overflow-y: auto;
  z-index: 9999;
}

.dropdown-item {
  padding: 8px 12px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.dropdown-item:hover {
  background-color: var(--gray-100);
}

.loading-indicator {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--gray-500);
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