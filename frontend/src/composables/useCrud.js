import { ref, onMounted, computed } from 'vue';
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
            items.value = Array.isArray(response.data) ? response.data : [];
        } catch (e) {
            error.value = `Не удалось загрузить данные для ресурса: ${resource}.`;
            console.error(e);
            items.value = []; // Ensure items is always an array
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
            const response = await apiClient.post(`/${resource}`, data);
            // Добавляем новый элемент в массив локально
            if (response.data && Array.isArray(items.value)) {
                items.value.push(response.data);
            } else {
                // Если сервер не вернул созданный объект, перезагружаем список
                await fetchItems();
            }
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
            const response = await apiClient.put(`/${resource}/${id}`, data);
            // Обновляем элемент в массиве локально
            if (response.data && Array.isArray(items.value)) {
                const index = items.value.findIndex(item => item.id === id);
                if (index !== -1) {
                    items.value[index] = response.data;
                } else {
                    // Если элемент не найден, перезагружаем список
                    await fetchItems();
                }
            } else {
                // Если сервер не вернул обновленный объект, перезагружаем список
                await fetchItems();
            }
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
            // Удаляем элемент из массива локально
            if (Array.isArray(items.value)) {
                const index = items.value.findIndex(item => item.id === id);
                if (index !== -1) {
                    items.value.splice(index, 1);
                }
            } else {
                // Если items не массив, перезагружаем данные
                await fetchItems();
            }
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

    // Ensure items is always an array when accessed
    const safeItems = computed(() => Array.isArray(items.value) ? items.value : []);

    // Возвращаем все переменные и функции, чтобы их можно было использовать в компонентах.
    return {
        items: safeItems,
        loading,
        error,
        fetchItems,
        createItem,
        updateItem,
        deleteItem,
    };
}