<script setup>
import { onMounted } from 'vue';
import useAuth from '../composables/useAuth';

const { user, fetchUser, logout } = useAuth();

// При монтировании компонента пытаемся загрузить данные пользователя
onMounted(async () => {
  await fetchUser();
});
</script>

<template>
  <div class="account-container">
    <div class="account-box">
      <h1>Личный кабинет</h1>
      <p>Добро пожаловать!</p>

      <!-- Этот блок будет работать, когда вы реализуете
           эндпоинт для получения данных пользователя -->
      <div v-if="user" class="user-info">
        <p><strong>Пользователь:</strong> {{ user.username }}</p>
        <p><strong>Роль:</strong> {{ user.role }}</p>
      </div>
      <div v-else>
        <p>Загрузка данных пользователя...</p>
      </div>

      <button @click="logout" class="logout-button">Выйти</button>
    </div>
  </div>
</template>

<style scoped>
.account-container {
  padding: 2rem;
  font-family: sans-serif;
}
.account-box {
  max-width: 800px;
  margin: 0 auto;
  padding: 2rem;
  background: #f9f9f9;
  border-radius: 8px;
}
.logout-button {
  margin-top: 1.5rem;
  padding: 0.5rem 1rem;
  background-color: #dc3545;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}
</style>