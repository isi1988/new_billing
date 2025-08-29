<script setup>
defineProps({
  isOpen: Boolean,
});
defineEmits(['close']);

const navLinks = [
  { name: 'Личный кабинет', path: '/account' },
  { name: 'Дашборд трафика', path: '/traffic' },
  { name: 'Клиенты', path: '/clients' },
  { name: 'Договоры', path: '/contracts' },
  { name: 'Подключения', path: '/connections' },
  { name: 'Тарифы', path: '/tariffs' },
  { name: 'Оборудование', path: '/equipment' },
  { name: 'Пользователи', path: '/users' },
  { name: 'Доработки', path: '/issues' },
];
</script>

<template>
  <!-- Overlay -->
  <div 
    v-if="isOpen" 
    class="drawer-overlay"
    @click="$emit('close')"
  />
  
  <!-- Drawer -->
  <aside 
    class="drawer"
    :class="{ 'drawer-open': isOpen, 'drawer-closed': !isOpen }"
  >
    <!-- Header -->
    <header class="drawer-header">
      <div class="drawer-header-content">
        <div class="drawer-logo">
          <div class="logo-icon">
            <svg class="logo-svg" fill="currentColor" viewBox="0 0 20 20">
              <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z"/>
            </svg>
          </div>
          <h2 class="logo-text">Ariadna Billing</h2>
        </div>
        <button 
          @click="$emit('close')" 
          class="close-button"
          aria-label="Закрыть меню"
        >
          <svg class="close-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>
    </header>
    
    <!-- Navigation -->
    <nav class="drawer-navigation">
      <div class="nav-container">
        <router-link 
          v-for="link in navLinks" 
          :key="link.path"
          :to="link.path" 
          @click="$emit('close')"
          class="nav-link"
        >
          <div class="nav-icon">
            <svg class="nav-svg" fill="currentColor" viewBox="0 0 20 20">
              <circle cx="10" cy="10" r="3"/>
            </svg>
          </div>
          <span class="nav-text">{{ link.name }}</span>
        </router-link>
      </div>
    </nav>
    
    <!-- Footer -->
    <footer class="drawer-footer">
      <div class="footer-text">
        Ariadna Billing System v1.0
      </div>
    </footer>
  </aside>
</template>

<style scoped>
/* Overlay */
.drawer-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 40;
  opacity: 1;
  transition: opacity 0.3s ease-in-out;
}

/* Drawer */
.drawer {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: 20rem; /* w-80 */
  background-color: white;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  z-index: 50;
  transform: translateX(100%);
  transition: transform 0.3s ease-in-out;
  display: flex;
  flex-direction: column;
}

.drawer-open {
  transform: translateX(0);
}

.drawer-closed {
  transform: translateX(100%);
}

/* Header */
.drawer-header {
  padding: 1rem 1.5rem;
  background: linear-gradient(135deg, var(--primary-500) 0%, var(--primary-600) 100%);
  color: white;
}

.drawer-header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.drawer-logo {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.logo-icon {
  width: 2rem;
  height: 2rem;
  background-color: white;
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.logo-svg {
  width: 1.25rem;
  height: 1.25rem;
  color: var(--primary-600);
}

.logo-text {
  font-size: 1.125rem;
  font-weight: 500;
  margin: 0;
}

.close-button {
  padding: 0.5rem;
  background-color: transparent;
  border: none;
  color: white;
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition: background-color 0.2s ease-in-out;
}

.close-button:hover {
  background-color: var(--primary-700);
}

.close-icon {
  width: 1.25rem;
  height: 1.25rem;
}

/* Navigation */
.drawer-navigation {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem 0;
}

.nav-container {
  padding: 0 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--gray-700);
  text-decoration: none;
  border-radius: var(--radius-xl);
  transition: all 0.2s ease-in-out;
  margin-bottom: 2px;
}

.nav-link:hover {
  background-color: var(--primary-50);
  color: var(--primary-700);
  text-decoration: none;
  transform: translateX(4px);
}

.nav-link.router-link-active {
  background-color: var(--primary-100);
  color: var(--primary-700);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  text-decoration: none;
  border-left: 4px solid var(--primary-600);
}

.nav-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  border-radius: var(--radius-lg);
  background-color: var(--gray-100);
  color: var(--gray-600);
  transition: all 0.2s ease-in-out;
}

.nav-link:hover .nav-icon {
  background-color: var(--primary-100);
  color: var(--primary-600);
}

.nav-link.router-link-active .nav-icon {
  background: linear-gradient(135deg, var(--primary-500) 0%, var(--primary-600) 100%);
  color: white;
  box-shadow: 0 2px 4px rgba(26, 115, 232, 0.2);
}

.nav-svg {
  width: 1.25rem;
  height: 1.25rem;
}

.nav-text {
  font-weight: 500;
}

/* Footer */
.drawer-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--gray-200);
  background-color: var(--gray-50);
}

.footer-text {
  font-size: 0.75rem;
  color: var(--gray-600);
  text-align: center;
  font-weight: 500;
}

/* Responsive */
@media (max-width: 640px) {
  .drawer {
    width: 100vw;
  }
}
</style>