import { defineStore } from "pinia";
import SandboxService from "@/services/sandboxService.js";

export const GeneralStore = defineStore("general", {
  state: () => ({
    showLoadingScreen: false, // Initialer Zustand
    sandboxes: [],
  }),
  getters: {
    isLoading() {
      return this.showLoadingScreen;
    },
    getSandboxes() {
      return this.sandboxes;
    },
  },
  actions: {
    setLoading(value) {
      this.showLoadingScreen = value;
    },
    addSandbox(sandbox) {
      // Überprüfen, ob die Sandbox schon existiert, um Duplikate zu vermeiden
      if (!this.sandboxes.some((env) => env.id === sandbox.id)) {
        this.sandboxes.push(sandbox);
      }
    },
    removeSandbox(sandboxId) {
      this.sandboxes = this.sandboxes.filter((env) => env.id !== sandboxId);
    },
    async refreshSandboxes() {
        for (let i =  0; i < this.sandboxes.length; i++) {
          let sandbox = this.sandboxes[i];
          const response = await SandboxService.refreshSandbox(sandbox);
          if (response.success) {
            this.sandboxes[i] = response.sandbox;
          } else {
            console.log("Failed to refresh sandbox, remove from list", response.message);
            this.removeSandbox(sandbox.id);
          }
        }
    }
  },
});
