<template>
  <header class="app-header">
    <h1 class="app-title">NewBilling</h1>
    <button @click="$emit('toggle-drawer')" class="drawer-toggle">
      ☰
    </button>
  </header>
</template>

<style scoped>
.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px;
  height: 64px;
  background-color: var(--surface-color);
  box-shadow: var(--shadow);
  color: var(--primary-color);
}
.app-title {
  font-size: 20px;
  font-weight: 500;
}
.drawer-toggle {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: var(--text-color);
}
</style>