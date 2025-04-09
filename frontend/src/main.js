import { createApp } from "vue";
import PrimeVue from "primevue/config";
import Aura from "@primevue/themes/aura";
import "./style.css";
import App from "./App.vue";
import router from "./router.js";
import "primeicons/primeicons.css";
import { createPinia } from "pinia";
import ToastService from "primevue/toastservice";
import { definePreset } from "@primevue/themes";
import piniaPluginPersistedState from "pinia-plugin-persistedstate"

const app = createApp(App);
const pinia = createPinia();
pinia.use(piniaPluginPersistedState)
app.use(pinia);
app.use(ToastService);
const MrPixelPreset = definePreset(Aura, {
  semantic: {
    primary: {
      50: '#effc00',
      100: '#e8f500',
      200: '#e1ed00',
      300: '#d7e300',
      400: '#cfdb00',
      500: '#c8d400', // mr. pixel green
      600: '#b2bd00',
      700: '#9fa800',
      800: '#8e9600',
      900: '#767d00',
      950: '#6a7000'
    },
    /*gray: {
      50: "{zinc.50}",
      100: "{zinc.100}",
      200: "{zinc.200}",
      300: "{zinc.300}",
      400: "{zinc.400}",
      500: "{zinc.500}",
      600: "{zinc.600}",
      700: "{zinc.700}",
      800: "{zinc.100}",
      900: "{zinc.900}",
      950: "{zinc.950}",
    },*/
  },
});
app.use(PrimeVue, {
  theme: {
    preset: MrPixelPreset, // Default: Aura
  },
});
app.use(router);
app.mount("#app");
