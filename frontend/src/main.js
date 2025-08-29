import { createApp } from 'vue'
import App from './App.vue'
import router from './router'

// Modern flat design styles
import './styles/global.css';

createApp(App)
    .use(router)
    .mount('#app')