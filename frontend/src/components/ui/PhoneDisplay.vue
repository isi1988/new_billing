<script setup>
const props = defineProps({
  phone: { type: String, required: true },
  format: { type: String, default: 'russian' }, // russian, international
});

function formatPhone(phoneNumber) {
  if (!phoneNumber) return '';
  
  // Убираем все нечисловые символы
  const digits = phoneNumber.replace(/\D/g, '');
  
  if (props.format === 'international') {
    // Международный формат +7 (XXX) XXX-XX-XX
    if (digits.length >= 11 && digits.startsWith('7')) {
      const formatted = digits.slice(1); // убираем первую 7
      return `+7 (${formatted.slice(0, 3)}) ${formatted.slice(3, 6)}-${formatted.slice(6, 8)}-${formatted.slice(8, 10)}`;
    }
    if (digits.length >= 10 && digits.startsWith('8')) {
      const formatted = digits.slice(1); // убираем первую 8
      return `+7 (${formatted.slice(0, 3)}) ${formatted.slice(3, 6)}-${formatted.slice(6, 8)}-${formatted.slice(8, 10)}`;
    }
  }
  
  // Российский формат 8 (XXX) XXX-XX-XX
  if (digits.length >= 11) {
    if (digits.startsWith('7')) {
      const formatted = digits.slice(1);
      return `8 (${formatted.slice(0, 3)}) ${formatted.slice(3, 6)}-${formatted.slice(6, 8)}-${formatted.slice(8, 10)}`;
    }
    if (digits.startsWith('8')) {
      const formatted = digits.slice(1);
      return `8 (${formatted.slice(0, 3)}) ${formatted.slice(3, 6)}-${formatted.slice(6, 8)}-${formatted.slice(8, 10)}`;
    }
  }
  
  if (digits.length >= 10) {
    return `8 (${digits.slice(0, 3)}) ${digits.slice(3, 6)}-${digits.slice(6, 8)}-${digits.slice(8, 10)}`;
  }
  
  // Если не подошел ни один формат, возвращаем как есть
  return phoneNumber;
}
</script>

<template>
  <span class="phone-display" :title="phone">
    {{ formatPhone(phone) }}
  </span>
</template>

<style scoped>
.phone-display {
  font-family: 'Courier New', monospace;
  font-weight: 500;
  color: var(--text-color);
}
</style>