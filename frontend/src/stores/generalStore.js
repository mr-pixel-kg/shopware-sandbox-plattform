import { defineStore } from "pinia";

export const GeneralStore = defineStore("general", {
  state: () => ({
    showLoadingScreen: false, // Initialer Zustand
    sandboxEnvironments: [],
  }),
  getters: {
    isLoading() {
      return this.showLoadingScreen;
    },
    getSandboxEnvironments() {
      return this.sandboxEnvironments;
    },
  },
  actions: {
    setLoading(value) {
      this.showLoadingScreen = value;
    },
    addSandbox(sandbox) {
      // Überprüfen, ob die Sandbox schon existiert, um Duplikate zu vermeiden
      if (
        !this.sandboxEnvironments.some(
          (env) => env.sandboxId === sandbox.sandboxId,
        )
      ) {
        this.sandboxEnvironments.push(sandbox);
      }
    },
    removeSandbox(sandboxId) {
      this.sandboxEnvironments = this.sandboxEnvironments.filter(
        (env) => env.sandboxId !== sandboxId,
      );
    },
  },
});
