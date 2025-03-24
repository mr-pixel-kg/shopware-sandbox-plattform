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

const app = createApp(App);
const pinia = createPinia();
app.use(pinia);
app.use(ToastService);
const MyPreset = definePreset(Aura, {
  semantic: {
    /*primary: {
      50: '{lime.50}',
      100: '{lime.100}',
      200: '{lime.200}',
      300: '{lime.300}',
      400: '{lime.400}',
      500: '{lime.500}',
      600: '{lime.600}',
      700: '{lime.700}',
      800: '{lime.800}',
      900: '{lime.900}',
      950: '{lime.950}'
    }
    dark: {
      50: "{zinc.50}",
      100: "{zinc.100}",
      200: "{zinc.200}",
      300: "{zinc.300}",
      400: "{zinc.400}",
      500: "{zinc.500}",
      600: "{zinc.600}",
      700: "{zinc.700}",
      800: "{zinc.800}",
      900: "{zinc.900}",
      950: "{zinc.950}",
    },*/
  },
});
app.use(PrimeVue, {
  theme: {
    preset: Aura,
  },
});
app.use(router);
app.mount("#app");
