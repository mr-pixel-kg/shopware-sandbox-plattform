import { computed, ref } from "vue";
import { defineStore } from "pinia";
import { api } from "@/lib/api";
import type { User } from "@/types/api";

const TOKEN_KEY = "shopshredder_auth_token";

export const useAuthStore = defineStore("auth", () => {
  const token = ref<string | null>(localStorage.getItem(TOKEN_KEY));
  const user = ref<User | null>(null);
  const ready = ref(false);

  const isAuthenticated = computed(() => Boolean(token.value && user.value));

  async function bootstrap() {
    if (!token.value) {
      ready.value = true;
      return;
    }

    try {
      user.value = await api.me(token.value);
    } catch {
      logout();
    } finally {
      ready.value = true;
    }
  }

  async function login(email: string, password: string) {
    const result = await api.login(email, password);
    token.value = result.token;
    user.value = result.user;
    localStorage.setItem(TOKEN_KEY, result.token);
  }

  function logout() {
    token.value = null;
    user.value = null;
    localStorage.removeItem(TOKEN_KEY);
  }

  return {
    token,
    user,
    ready,
    isAuthenticated,
    bootstrap,
    login,
    logout,
  };
});
