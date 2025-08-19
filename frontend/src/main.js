import { createApp } from 'vue'
import App from './App.vue'
import router from './router'

// УБЕДИТЕСЬ, ЧТО ВСЕ ТРИ ФАЙЛА ИМПОРТИРУЮТСЯ ЗДЕСЬ
import './assets/css/main.css';
import './assets/css/components.css';
import './assets/css/forms.css';

createApp(App)
    .use(router)
    .mount('#app')