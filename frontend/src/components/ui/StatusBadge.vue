<script setup>
const props = defineProps({
  type: { type: String, required: true }, // client_type, payment_type, status, etc.
  value: { type: String, required: true },
  size: { type: String, default: 'medium' }, // small, medium, large
});

function getBadgeClass() {
  const baseClass = 'status-badge';
  const sizeClass = `status-badge--${props.size}`;
  
  let variantClass = 'status-badge--default';
  
  switch (props.type) {
    case 'client_type':
      variantClass = props.value === 'individual' ? 'status-badge--individual' : 'status-badge--legal';
      break;
    case 'payment_type':
      variantClass = props.value === 'prepaid' ? 'status-badge--prepaid' : 'status-badge--postpaid';
      break;
    case 'issue_status':
      variantClass = props.value === 'new' ? 'status-badge--new' : 'status-badge--resolved';
      break;
    case 'blocked_status':
      variantClass = props.value ? 'status-badge--blocked' : 'status-badge--active';
      break;
    default:
      variantClass = 'status-badge--default';
  }
  
  return `${baseClass} ${sizeClass} ${variantClass}`;
}

function getDisplayText() {
  switch (props.type) {
    case 'client_type':
      return props.value === 'individual' ? 'Физ. лицо' : 'Юр. лицо';
    case 'payment_type':
      return props.value === 'prepaid' ? 'Предоплата' : 'Постоплата';
    case 'issue_status':
      return props.value === 'new' ? 'Новая' : 'Решена';
    case 'blocked_status':
      return props.value ? 'Заблокирован' : 'Активен';
    default:
      return props.value;
  }
}
</script>

<template>
  <span :class="getBadgeClass()">
    {{ getDisplayText() }}
  </span>
</template>

<style scoped>
.status-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.025em;
  border-radius: 9999px;
  white-space: nowrap;
  transition: all 0.2s ease;
}

/* Sizes */
.status-badge--small {
  padding: 0.125rem 0.5rem;
  font-size: 0.625rem;
  line-height: 1;
}

.status-badge--medium {
  padding: 0.25rem 0.75rem;
  font-size: 0.75rem;
  line-height: 1.25;
}

.status-badge--large {
  padding: 0.375rem 1rem;
  font-size: 0.875rem;
  line-height: 1.5;
}

/* Variants */
.status-badge--individual {
  background-color: #e0f2fe;
  color: #0369a1;
  border: 1px solid #bae6fd;
}

.status-badge--legal {
  background-color: #fef3c7;
  color: #d97706;
  border: 1px solid #fed7aa;
}

.status-badge--prepaid {
  background-color: #dcfce7;
  color: #166534;
  border: 1px solid #bbf7d0;
}

.status-badge--postpaid {
  background-color: #fdf2f8;
  color: #be185d;
  border: 1px solid #fbcfe8;
}

.status-badge--new {
  background-color: #fef3c7;
  color: #d97706;
  border: 1px solid #fed7aa;
}

.status-badge--resolved {
  background-color: #dcfce7;
  color: #166534;
  border: 1px solid #bbf7d0;
}

.status-badge--blocked {
  background-color: #fee2e2;
  color: #dc2626;
  border: 1px solid #fecaca;
}

.status-badge--active {
  background-color: #dcfce7;
  color: #166534;
  border: 1px solid #bbf7d0;
}

.status-badge--default {
  background-color: #f1f5f9;
  color: #475569;
  border: 1px solid #e2e8f0;
}

/* Hover effects */
.status-badge:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}
</style>