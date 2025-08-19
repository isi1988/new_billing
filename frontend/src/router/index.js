import { createRouter, createWebHistory } from 'vue-router';

// --- Импорт страниц (Views) ---
import LoginView from '../views/LoginView.vue';
import AccountView from '../views/AccountView.vue';
import UserManagement from '../views/UserManagement.vue';
import EquipmentManagement from '../views/EquipmentManagement.vue';
import TariffManagement from '../views/TariffManagement.vue';
import ConnectionManagement from '../views/ConnectionManagement.vue';
import ContractManagement from '../views/ContractManagement.vue';
import ClientManagement from "../views/ClientManagement.vue";
// Задел на будущее: импортируем страницу клиентов, которую создадим следующей
// import ClientManagement from '../views/ClientManagement.vue';

const routes = [
    // --- Основные маршруты ---
    {
        // При заходе на корень сайта, перенаправляем в личный кабинет
        path: '/',
        redirect: '/account',
    },
    {
        // Страница входа в систему (публичная)
        path: '/login',
        name: 'Login',
        component: LoginView,
    },
    {
        // Личный кабинет / Главная страница после входа (защищенная)
        path: '/account',
        name: 'Account',
        component: AccountView,
        meta: { requiresAuth: true }, // Эта страница требует аутентификации
    },

    // --- Маршруты для управления сущностями (все защищенные) ---
    {
        path: '/users',
        name: 'UserManagement',
        component: UserManagement,
        meta: { requiresAuth: true },
    },
    {
        path: '/tariffs',
        name: 'TariffManagement',
        component: TariffManagement,
        meta: { requiresAuth: true },
    },
    {
        path: '/equipment',
        name: 'EquipmentManagement',
        component: EquipmentManagement,
        meta: { requiresAuth: true },
    },
    {
        path: '/contracts',
        name: 'ContractManagement',
        component: ContractManagement,
        meta: { requiresAuth: true },
    },
    {
        path: '/connections',
        name: 'ConnectionManagement',
        component: ConnectionManagement,
        meta: { requiresAuth: true },
    },
    {
      // Маршрут для клиентов (когда вы создадите ClientManagement.vue)
      path: '/clients',
      name: 'ClientManagement',
      component: ClientManagement,
      meta: { requiresAuth: true },
    },
];

const router = createRouter({
    history: createWebHistory(),
    routes,
});

// --- Навигационный страж (Navigation Guard) ---
// Эта функция выполняется перед каждым переходом по маршруту.
// Она проверяет, есть ли у пользователя доступ к запрашиваемой странице.
router.beforeEach((to, from, next) => {
    // Проверяем, есть ли токен в локальном хранилище
    const token = localStorage.getItem('token');

    // 1. Если маршрут требует аутентификации (`meta.requiresAuth`) и у пользователя НЕТ токена...
    if (to.meta.requiresAuth && !token) {
        // ...то перенаправляем его на страницу входа.
        next({ name: 'Login' });
    }
    // 2. Если пользователь уже авторизован (есть токен) и пытается зайти на страницу входа...
    else if (to.name === 'Login' && token) {
        // ...то перенаправляем его в личный кабинет, чтобы он не логинился заново.
        next({ name: 'Account' });
    }
    // 3. Во всех остальных случаях...
    else {
        // ...разрешаем переход.
        next();
    }
});

export default router;