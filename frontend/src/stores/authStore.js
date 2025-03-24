import { defineStore } from "pinia";
import ApiService from "@/services/apiService";

export const useAuthStore = defineStore("auth", {
    state: () => ({
        username: null,
        password: null,
        isAuthenticated: false,
    }),
    actions: {
        async login(username, password) {
            const success = await ApiService.login(username, password);
            if (success) {
                this.username = username;
                this.password = password;
                this.isAuthenticated = true;
            }
            return success;
        },

        logout() {
            this.username = null;
            this.password = null;
            this.isAuthenticated = false;
            ApiService.logout();
        },
    },
    persist: true,
});