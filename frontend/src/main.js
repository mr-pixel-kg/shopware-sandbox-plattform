import { createApp } from 'vue';
import PrimeVue from 'primevue/config';
import Aura from '@primevue/themes/aura';
import './style.css'
import App from './App.vue'
import router from "./router.js";
import 'primeicons/primeicons.css'
import {createPinia} from "pinia";
import ToastService from 'primevue/toastservice';


const app = createApp(App);
const pinia = createPinia()
app.use(pinia)
app.use(ToastService);
app.use(PrimeVue, {
    theme: {
        preset: Aura
    }
});
app.use(router)
app.mount('#app');
