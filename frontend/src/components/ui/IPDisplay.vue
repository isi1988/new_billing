<script setup>
import { ref, onMounted } from 'vue';
import apiClient from '@/api/client';

const props = defineProps({
  ip: { type: String, required: true },
  showPopup: { type: Boolean, default: true }
});

const ipInfo = ref(null);
const loading = ref(false);
const showTooltip = ref(false);
const tooltipPosition = ref({ x: 0, y: 0 });

async function fetchIPInfo() {
  if (!props.showPopup || ipInfo.value || loading.value) return;
  
  loading.value = true;
  try {
    const response = await apiClient.get(`/ip/${props.ip}/info`);
    ipInfo.value = response.data;
  } catch (error) {
    console.error('Ошибка получения информации об IP:', error);
  } finally {
    loading.value = false;
  }
}

function handleMouseEnter(event) {
  if (!props.showPopup) return;
  
  showTooltip.value = true;
  updateTooltipPosition(event);
  fetchIPInfo();
}

function handleMouseLeave() {
  showTooltip.value = false;
}

function updateTooltipPosition(event) {
  const rect = event.target.getBoundingClientRect();
  tooltipPosition.value = {
    x: rect.left + rect.width / 2,
    y: rect.bottom + 5
  };
}

function isPrivateIP(ip) {
  const parts = ip.split('.').map(Number);
  if (parts.length !== 4) return false;
  
  // 10.0.0.0/8
  if (parts[0] === 10) return true;
  
  // 172.16.0.0/12
  if (parts[0] === 172 && parts[1] >= 16 && parts[1] <= 31) return true;
  
  // 192.168.0.0/16
  if (parts[0] === 192 && parts[1] === 168) return true;
  
  // 127.0.0.0/8 (localhost)
  if (parts[0] === 127) return true;
  
  return false;
}

function formatDate(dateString) {
  if (!dateString) return 'Неизвестно';
  return new Date(dateString).toLocaleString('ru-RU');
}
</script>

<template>
  <span 
    class="ip-display"
    :class="{ 'private-ip': isPrivateIP(ip), 'public-ip': !isPrivateIP(ip) }"
    @mouseenter="handleMouseEnter"
    @mouseleave="handleMouseLeave"
  >
    {{ ip }}
    
    <!-- Tooltip -->
    <div 
      v-if="showTooltip && showPopup" 
      class="ip-tooltip"
      :style="{
        left: tooltipPosition.x + 'px',
        top: tooltipPosition.y + 'px'
      }"
    >
      <div v-if="loading" class="tooltip-loading">
        <span class="material-icons icon-sm spinning">hourglass_empty</span>
        Загрузка...
      </div>
      
      <div v-else-if="ipInfo" class="tooltip-content">
        <div class="tooltip-header">
          <strong>{{ ip }}</strong>
          <span class="ip-type">{{ isPrivateIP(ip) ? 'Внутренний' : 'Внешний' }}</span>
        </div>
        
        <!-- Hostname information -->
        <div v-if="ipInfo.hostname" class="tooltip-section">
          <div class="tooltip-label">Имя хоста:</div>
          <div class="tooltip-value">{{ ipInfo.hostname.hostname }}</div>
          <div class="tooltip-date">
            Обновлено: {{ formatDate(ipInfo.hostname.updated_at) }}
          </div>
        </div>
        
        <!-- Connection information -->
        <div v-if="ipInfo.connection" class="tooltip-section">
          <div class="tooltip-label">Подключение:</div>
          <div class="connection-info">
            <div><strong>Адрес:</strong> {{ ipInfo.connection.address }}</div>
            <div><strong>Тариф:</strong> {{ ipInfo.connection.tariff_name }}</div>
            <div><strong>Клиент:</strong> {{ ipInfo.connection.client_name }}</div>
            <div><strong>Договор:</strong> {{ ipInfo.connection.contract_number }}</div>
          </div>
        </div>
        
        <div v-if="!ipInfo.hostname && !ipInfo.connection" class="tooltip-section">
          <div class="no-info">Дополнительная информация отсутствует</div>
        </div>
      </div>
      
      <div v-else class="tooltip-content">
        <div class="tooltip-header">
          <strong>{{ ip }}</strong>
          <span class="ip-type">{{ isPrivateIP(ip) ? 'Внутренний' : 'Внешний' }}</span>
        </div>
        <div class="no-info">Информация не найдена</div>
      </div>
    </div>
  </span>
</template>

<style scoped>
.ip-display {
  position: relative;
  cursor: pointer;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 0.875rem;
  padding: 0.25rem 0.5rem;
  border-radius: var(--radius-md);
  transition: all 0.2s ease;
}

.private-ip {
  background-color: var(--blue-50);
  color: var(--blue-700);
  border: 1px solid var(--blue-200);
}

.private-ip:hover {
  background-color: var(--blue-100);
  border-color: var(--blue-300);
}

.public-ip {
  background-color: var(--orange-50);
  color: var(--orange-700);
  border: 1px solid var(--orange-200);
}

.public-ip:hover {
  background-color: var(--orange-100);
  border-color: var(--orange-300);
}

.ip-tooltip {
  position: fixed;
  z-index: 9999;
  background: white;
  border: 1px solid var(--gray-300);
  border-radius: var(--radius-lg);
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.15);
  padding: 1rem;
  min-width: 280px;
  max-width: 400px;
  transform: translateX(-50%);
  animation: fadeIn 0.2s ease-in-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateX(-50%) translateY(-5px); }
  to { opacity: 1; transform: translateX(-50%) translateY(0); }
}

.tooltip-loading {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--gray-600);
  font-size: 0.875rem;
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.tooltip-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--gray-200);
}

.ip-type {
  font-size: 0.75rem;
  padding: 0.25rem 0.5rem;
  border-radius: var(--radius-full);
  font-weight: 500;
}

.private-ip .ip-type {
  background-color: var(--blue-100);
  color: var(--blue-700);
}

.public-ip .ip-type {
  background-color: var(--orange-100);
  color: var(--orange-700);
}

.tooltip-section {
  margin-bottom: 1rem;
}

.tooltip-section:last-child {
  margin-bottom: 0;
}

.tooltip-label {
  font-weight: 600;
  color: var(--gray-800);
  margin-bottom: 0.25rem;
  font-size: 0.875rem;
}

.tooltip-value {
  color: var(--gray-700);
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 0.875rem;
}

.tooltip-date {
  color: var(--gray-500);
  font-size: 0.75rem;
  margin-top: 0.25rem;
}

.connection-info {
  font-size: 0.875rem;
  line-height: 1.4;
}

.connection-info > div {
  margin-bottom: 0.25rem;
  color: var(--gray-700);
}

.connection-info strong {
  color: var(--gray-800);
}

.no-info {
  color: var(--gray-500);
  font-style: italic;
  font-size: 0.875rem;
  text-align: center;
  padding: 0.5rem 0;
}
</style>