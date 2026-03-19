import { createApp } from 'vue'
import pinia from './stores'
import { useAuthStore } from './stores/auth.store'
import router from './router'
import App from './App.vue'
import './style.css'

async function bootstrap() {
  const app = createApp(App)
  app.use(pinia)

  const authStore = useAuthStore()
  await authStore.initialize()

  app.use(router)
  app.mount('#app')
}

bootstrap()
