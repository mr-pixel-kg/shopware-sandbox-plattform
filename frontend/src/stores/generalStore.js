import { defineStore } from 'pinia';

export const GeneralStore = defineStore('general', {
    state: () => ({
        showLoadingScreen: false,  // Initialer Zustand
    }),
    getters: {
        isLoading() {
            return this.showLoadingScreen;
        }
    },
    actions: {
        setLoading(value) {
            this.showLoadingScreen = value;
        }
    }
});