import { reactive, computed } from 'vue';
import { useRouter } from 'vue-router';
import apiClient from '../api/client';

// Глобальное реактивное состояние для аутентификации
const state = reactive({
    token: localStorage.getItem('token') || null,
    user: null, // Здесь можно хранить данные пользователя
});

export default function useAuth() {
    const router = useRouter();

    const isAuthenticated = computed(() => !!state.token);
    const user = computed(() => state.user);

    const login = async (username, password) => {
        try {
            const response = await apiClient.post('/login', { username, password });
            const token = response.data.token;

            // Сохраняем токен
            localStorage.setItem('token', token);
            state.token = token;

            // Опционально: можно загрузить данные пользователя
            // await fetchUser();

            // Перенаправляем в личный кабинет
            await router.push('/account');
        } catch (error) {
            console.error("Ошибка входа:", error);
            // Очищаем токен на случай неудачной попытки
            localStorage.removeItem('token');
            state.token = null;
            throw error; // Пробрасываем ошибку дальше, чтобы обработать в компоненте
        }
    };

    const logout = () => {
        localStorage.removeItem('token');
        state.token = null;
        state.user = null;
        router.push('/login');
    };

    // Пример функции для получения данных о текущем пользователе
    // (требует эндпоинт на бэкенде, например, /api/users/me)
    const fetchUser = async () => {
        if (!state.token) return;
        try {
            // Пока что у нас нет эндпоинта /api/users/me,
            // поэтому эта часть - задел на будущее.
            // const response = await apiClient.get('/users/me');
            // state.user = response.data;
        } catch (error) {
            console.error("Не удалось получить данные пользователя:", error);
            // Если токен невалидный, выходим из системы
            if (error.response && error.response.status === 401) {
                logout();
            }
        }
    };

    return {
        isAuthenticated,
        user,
        login,
        logout,
        fetchUser,
    };
}