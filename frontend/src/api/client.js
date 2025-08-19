import axios from 'axios';

const apiClient = axios.create({
    // Vite proxy перенаправит /api на ваш Go бэкенд
    baseURL: '/api',
});

// Используем interceptor для добавления токена в каждый запрос
apiClient.interceptors.request.use(config => {
    const token = localStorage.getItem('token');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
}, error => {
    return Promise.reject(error);
});

export default apiClient;