<script setup>
defineProps({
  isOpen: Boolean,
});
defineEmits(['close']);

const navLinks = [
  { name: 'Личный кабинет', path: '/account' },
  { name: 'Клиенты', path: '/clients' },
  { name: 'Договоры', path: '/contracts' },
  { name: 'Подключения', path: '/connections' },
  { name: 'Тарифы', path: '/tariffs' },
  { name: 'Оборудование', path: '/equipment' },
  { name: 'Пользователи', path: '/users' },
];
</script>

<template>
  <!-- Полупрозрачный фон, который появляется при открытии меню -->
  <div class="drawer-overlay" :class="{ open: isOpen }" @click="$emit('close')">
    <!-- Само меню, которое выезжает -->
    <aside class="drawer-content" :class="{ open: isOpen }" @click.stop>
      <header class="drawer-header">
        <h2>Меню</h2>
        <button @click="$emit('close')" class="close-button">&times;</button>
      </header>
      <nav class="drawer-nav">
        <ul>
          <li v-for="link in navLinks" :key="link.path">
            <router-link :to="link.path" @click="$emit('close')">{{ link.name }}</router-link>
          </li>
        </ul>
      </nav>
    </aside>
  </div>
</template>

<style scoped>
.drawer-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 1000;
  opacity: 0;
  visibility: hidden;
  transition: opacity 0.3s ease, visibility 0.3s ease;
}
.drawer-overlay.open {
  opacity: 1;
  visibility: visible;
}

.drawer-content {
  position: fixed;
  top: 0;
  right: 0; /* Меню будет справа */
  bottom: 0;
  width: 300px;
  background-color: var(--surface-color);
  box-shadow: -2px 0 8px rgba(0,0,0,0.15);
  transform: translateX(100%); /* Изначально скрыто за правым краем */
  transition: transform 0.3s ease;
  z-index: 1001;
  display: flex;
  flex-direction: column;
}
.drawer-content.open {
  transform: translateX(0); /* Выезжает в видимую область */
}

.drawer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-color);
}
.drawer-header h2 {
  margin: 0;
  font-size: 20px;
  color: var(--text-color);
}
.close-button {
  background: none;
  border: none;
  font-size: 28px;
  cursor: pointer;
  color: var(--text-color-light);
}

.drawer-nav {
  padding: 16px 0;
}
.drawer-nav ul {
  list-style: none;
  padding: 0;
  margin: 0;
}
.drawer-nav a {
  display: block;
  padding: 12px 24px;
  text-decoration: none;
  color: var(--text-color);
  font-weight: 500;
  transition: background-color 0.2s;
}
.drawer-nav a:hover {
  background-color: #f0f2f5;
}
/* Стиль для активной ссылки */
.drawer-nav a.router-link-active {
  color: var(--primary-color);
  background-color: #e9e8f8;
}
</style>