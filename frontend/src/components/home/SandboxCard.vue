<script>
import Card from "primevue/card";
import Button from "primevue/button";
import { SandboxEnvironment } from "@/models/SandboxEnvironment.js";
import SandboxService from "@/services/sandboxService.js";
import { GeneralStore } from "@/stores/generalStore.js";

export default {
  name: "SandboxCard",

  components: {
    Card,
    Button,
  },

  props: {
    sandboxEnvironment: SandboxEnvironment,
  },

  computed: {
    sandbox() {
      return this.sandboxEnvironment;
    },
  },

  setup() {
    const store = GeneralStore();
    return {
      generalStore: store,
    };
  },

  methods: {
    openUrl() {
      window.open("https://" + this.sandbox.url, "_blank").focus();
    },

    onDelete() {
      try {
        const response = SandboxService.deleteSandbox(this.sandbox.sandboxId);
        console.log("Sandbox deleted", response);
        this.generalStore.removeSandbox(this.sandbox.sandboxId);
      } catch (e) {
        console.log("Failed to delete sandbox", e);
      }
    },
  },
};
</script>

<template>
  <Card style="width: 25rem; overflow: hidden">
    <template #title>Shopware Sandbox</template>
    <template #subtitle>{{ sandbox.image }}</template>
    <template #footer>
      <div class="flex w-full gap-2 mt-1">
        <Button
          icon="pi pi-trash"
          rounded
          severity="danger"
          aria-label="Cancel"
          class="w-1/3"
          @click="onDelete"
        />
        <Button
          severity="primary"
          aria-label="Cancel"
          class="w-full"
          label="Öffnen"
          @click="openUrl"
        />
      </div>
    </template>
  </Card>
</template>

<style scoped></style>
