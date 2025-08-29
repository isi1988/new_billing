<template>
  <header class="app-header">
    <div class="header-container">
      <div class="header-content">
        <div class="header-brand">
          <div class="brand-logo">
            <svg class="logo-svg" fill="currentColor" viewBox="0 0 20 20">
              <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z"/>
            </svg>
          </div>
          <div class="brand-text">
            <h1 class="brand-title">Ariadna Billing</h1>
            <p class="brand-subtitle">Система управления биллингом</p>
          </div>
        </div>
        
        <div class="header-actions">
          <div class="system-status">
            <div class="status-indicator"></div>
            <span class="status-text">Система активна</span>
          </div>
          
          <button 
            @click="$emit('toggle-drawer')" 
            class="menu-button"
            aria-label="Открыть меню"
          >
            <svg class="menu-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"/>
            </svg>
          </button>
        </div>
      </div>
    </div>
  </header>
</template>

<style scoped>
.app-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 50;
  background: white;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  border-bottom: 1px solid var(--gray-200);
}

.header-container {
  padding: 0.75rem 1.5rem;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-brand {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.brand-logo {
  width: 2.5rem;
  height: 2.5rem;
  background: linear-gradient(135deg, var(--primary-500) 0%, var(--primary-600) 100%);
  border-radius: var(--radius-xl);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 4px rgba(26, 115, 232, 0.2);
}

.logo-svg {
  width: 1.5rem;
  height: 1.5rem;
  color: white;
}

.brand-text {
  display: flex;
  flex-direction: column;
}

.brand-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--gray-900);
  margin: 0;
  line-height: 1.2;
}

.brand-subtitle {
  font-size: 0.75rem;
  color: var(--gray-500);
  font-weight: 500;
  margin: 0;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.system-status {
  display: none;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  background-color: var(--success-50);
  border: 1px solid var(--success-200);
  border-radius: var(--radius-full);
}

.status-indicator {
  width: 0.5rem;
  height: 0.5rem;
  background-color: var(--success-500);
  border-radius: 50%;
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

.status-text {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--success-700);
}

.menu-button {
  padding: 0.75rem;
  color: var(--gray-600);
  background: transparent;
  border: none;
  border-radius: var(--radius-xl);
  cursor: pointer;
  transition: all 0.2s ease-in-out;
}

.menu-button:hover {
  color: var(--gray-900);
  background-color: var(--gray-100);
}

.menu-icon {
  width: 1.5rem;
  height: 1.5rem;
}

@media (min-width: 768px) {
  .system-status {
    display: flex;
  }
}

@media (max-width: 640px) {
  .header-container {
    padding: 0.5rem 1rem;
  }
  
  .header-brand {
    gap: 0.75rem;
  }
  
  .brand-logo {
    width: 2rem;
    height: 2rem;
  }
  
  .logo-svg {
    width: 1.25rem;
    height: 1.25rem;
  }
  
  .brand-title {
    font-size: 1.125rem;
  }
  
  .brand-subtitle {
    display: none;
  }
}
</style>