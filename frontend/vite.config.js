import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path' // 1. Импортируем модуль 'path' из Node.js

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],

  // --- Настройка прокси для разработки ---
  server: {
    port: 5173, // Можете указать любой удобный порт
    proxy: {
      // Все запросы, начинающиеся с /api, перенаправляем на Go бэкенд
      '/api': {
        target: 'http://localhost:8080', // Убедитесь, что порт совпадает с вашим config.yaml
        changeOrigin: true,
      }
    }
  },

  // --- ДОБАВЛЕНА СЕКЦИЯ ДЛЯ ПСЕВДОНИМОВ ---
  resolve: {
    alias: {
      // Здесь мы говорим Vite: "когда видишь '@' в пути импорта,
      // замени его на абсолютный путь к папке 'frontend/src'".
      '@': path.resolve(__dirname, './src'),
    }
  }
})