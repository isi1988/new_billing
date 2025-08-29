import axios from 'axios';
import { useNotificationStore } from '@/stores/notification';

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

// Response interceptor для обработки ошибок
apiClient.interceptors.response.use(
    response => response,
    error => {
        // Получаем store для уведомлений
        const notificationStore = useNotificationStore();
        
        let errorTitle = 'Ошибка сети';
        let errorMessage = 'Произошла ошибка при обращении к серверу';
        let errorDetails = '';

        // Собираем детальную информацию об ошибке
        if (error.response) {
            // Ошибка от сервера
            errorTitle = `Ошибка ${error.response.status}`;
            errorMessage = error.response.data?.message || error.response.statusText || 'Ошибка сервера';
            
            errorDetails = `URL: ${error.config?.url}\n`;
            errorDetails += `Method: ${error.config?.method?.toUpperCase()}\n`;
            errorDetails += `Status: ${error.response.status}\n`;
            errorDetails += `Response: ${JSON.stringify(error.response.data, null, 2)}`;
            
        } else if (error.request) {
            // Ошибка сети (нет ответа)
            errorTitle = 'Нет связи с сервером';
            errorMessage = 'Не удалось подключиться к серверу. Проверьте соединение с интернетом.';
            
            errorDetails = `URL: ${error.config?.url}\n`;
            errorDetails += `Method: ${error.config?.method?.toUpperCase()}\n`;
            errorDetails += 'Status: No response from server\n';
            errorDetails += `Error: ${error.message}`;
            
        } else {
            // Другие ошибки
            errorMessage = error.message;
            errorDetails = `Error: ${error.message}\nConfig: ${JSON.stringify(error.config, null, 2)}`;
        }

        // Выводим в консоль
        console.error('API Error:', {
            title: errorTitle,
            message: errorMessage,
            details: errorDetails,
            originalError: error
        });

        // Показываем уведомление
        notificationStore.showError(errorTitle, errorMessage, errorDetails);

        return Promise.reject(error);
    }
);

export default apiClient;