<script setup>
import { ref } from 'vue';
import useAuth from '../composables/useAuth';

const { login } = useAuth();

const username = ref('');
const password = ref('');
const errorMessage = ref(null);
const isLoading = ref(false);

const handleLogin = async () => {
  errorMessage.value = null;
  isLoading.value = true;
  try {
    await login(username.value, password.value);
    // Роутер сам сделает редирект после успешного логина
  } catch (error) {
    errorMessage.value = 'Неверное имя пользователя или пароль.';
  } finally {
    isLoading.value = false;
  }
};
</script>

<template>
  <div class="login-container">
    <div class="login-box">
      <h1>Вход в систему</h1>
      <form @submit.prevent="handleLogin">
        <div class="input-group">
          <label for="username">Имя пользователя</label>
          <input id="username" type="text" v-model="username" required />
        </div>
        <div class="input-group">
          <label for="password">Пароль</label>
          <input id="password" type="password" v-model="password" required />
        </div>
        <div v-if="errorMessage" class="error-message">
          {{ errorMessage }}
        </div>
        <button type="submit" :disabled="isLoading">
          {{ isLoading ? 'Вход...' : 'Войти' }}
        </button>
      </form>
    </div>
  </div>
</template>

<style scoped>
/* Стили добавлены для наглядности */
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background-color: #f0f2f5;
}
.login-box {
  padding: 2rem;
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  width: 100%;
  max-width: 400px;
}
h1 {
  text-align: center;
  margin-bottom: 1.5rem;
}
.input-group {
  margin-bottom: 1rem;
}
.input-group label {
  display: block;
  margin-bottom: 0.5rem;
}
.input-group input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #ccc;
  border-radius: 4px;
}
button {
  width: 100%;
  padding: 0.75rem;
  border: none;
  background-color: #007bff;
  color: white;
  border-radius: 4px;
  font-size: 1rem;
  cursor: pointer;
}
button:disabled {
  background-color: #a0cfff;
}
.error-message {
  color: red;
  margin-bottom: 1rem;
  text-align: center;
}
</style>