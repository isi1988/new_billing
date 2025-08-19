import { ref, onMounted } from 'vue';
import apiClient from '../api/client';

/**
 * Универсальный composable для CRUD-операций.
 * @param {string} resource - Название ресурса API (например, 'users', 'tariffs').
 * @returns {object} - Реактивные переменные и функции для управления CRUD.
 */
export function useCrud(resource) {
    // --- РЕАКТИВНОЕ СОСТОЯНИЕ ---

    // Массив для хранения списка элементов
    const items = ref([]);
    // Флаг, указывающий, идет ли в данный момент загрузка данных
    const loading = ref(false);
    // Переменная для хранения текста ошибки
    const error = ref(null);

    // --- CRUD ФУНКЦИИ ---

    /**
     * READ (List)
     * Запрашивает и получает список всех элементов с сервера.
     */
    const fetchItems = async () => {
        loading.value = true;
        error.value = null;
        try {
            const response = await apiClient.get(`/${resource}`);
            // Присваиваем полученные данные или пустой массив, если данных нет
            items.value = response.data || [];
        } catch (e) {
            error.value = `Не удалось загрузить данные для ресурса: ${resource}.`;
            console.error(e);
        } finally {
            // Вне зависимости от результата, убираем состояние загрузки
            loading.value = false;
        }
    };

    /**
     * CREATE
     * Отправляет данные нового элемента на сервер.
     * @param {object} data - Объект с данными для создания.
     */
    const createItem = async (data) => {
        loading.value = true;
        error.value = null;
        try {
            await apiClient.post(`/${resource}`, data);
            await fetchItems(); // Обновляем список после успешного создания
        } catch (e) {
            error.value = `Ошибка при создании элемента.`;
            console.error(e);
            throw e; // Пробрасываем ошибку выше, чтобы компонент мог ее обработать
        } finally {
            loading.value = false;
        }
    };

    /**
     * UPDATE
     * Отправляет обновленные данные элемента на сервер.
     * @param {number|string} id - ID элемента для обновления.
     * @param {object} data - Объект с обновленными данными.
     */
    const updateItem = async (id, data) => {
        loading.value = true;
        error.value = null;
        try {
            await apiClient.put(`/${resource}/${id}`, data);
            await fetchItems(); // Обновляем список после успешного обновления
        } catch (e) {
            error.value = `Ошибка при обновлении элемента с ID: ${id}.`;
            console.error(e);
            throw e;
        } finally {
            loading.value = false;
        }
    };

    /**
     * DELETE
     * Отправляет запрос на удаление элемента с сервера.
     * @param {number|string} id - ID элемента для удаления.
     */
    const deleteItem = async (id) => {
        loading.value = true;
        error.value = null;
        try {
            await apiClient.delete(`/${resource}/${id}`);
            await fetchItems(); // Обновляем список после успешного удаления
        } catch (e) {
            error.value = `Ошибка при удалении элемента с ID: ${id}.`;
            console.error(e);
            throw e;
        } finally {
            loading.value = false;
        }
    };

    // --- ЖИЗНЕННЫЙ ЦИКЛ ---

    // Вызываем `fetchItems` один раз, когда компонент, использующий этот composable,
    // будет примонтирован к DOM.
    onMounted(fetchItems);

    // --- ВОЗВРАЩАЕМЫЕ ЗНАЧЕНИЯ ---

    // Возвращаем все переменные и функции, чтобы их можно было использовать в компонентах.
    return {
        items,
        loading,
        error,
        fetchItems,
        createItem,
        updateItem,
        deleteItem,
    };
}