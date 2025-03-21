<script>
import Card from "primevue/card";
import DataTable from "primevue/datatable";
import Column from "primevue/column";
import Button from "primevue/button";
import Tag from "primevue/tag";
import Dialog from "primevue/dialog";
import InputText from "primevue/inputtext";
import ProgressSpinner from "primevue/progressspinner";
import sandboxService from "@/services/sandboxService.js";
import { GeneralStore } from "@/stores/generalStore";
import { SandboxEnvironment } from "@/models/SandboxEnvironment.js";
import { SandboxImage } from "@/models/SandboxImage.js";

export default {
  components: {
    Card,
    DataTable,
    Column,
    Button,
    Tag,
    Dialog,
    InputText,
    ProgressSpinner,
  },

  props: {
    sandboxImage: SandboxImage,
  },

  computed: {
    image() {
      return this.sandboxImage;
    },
  },

  setup() {
    const store = GeneralStore();
    return {
      generalStore: store,
    };
  },

  methods: {
    async createSandbox() {
      this.generalStore.setLoading(true);
      console.log("Start loading");

      const startTime = Date.now();

      try {
        const response = await sandboxService.createSandbox(
          this.sandboxImage.imageName,
          60,
        );

        if (response.status === "success") {
          console.log("Sandbox erfolgreich erstellt", response);

          this.$toast.add({
            severity: "success",
            summary: this.sandboxImage.title,
            detail: "Sandbox erfolgreich erstellt!",
            life: 3000,
          });

          const sandboxEnvironment = new SandboxEnvironment(
            response.sandbox_id,
            response.image,
            response.url,
          );
          this.generalStore.addSandbox(sandboxEnvironment);
        } else {
          const errorMessage = response.message || "Unbekannter Fehler";
          console.log("Fehler beim Erstellen der Sandbox:", errorMessage);
          this.$toast.add({
            severity: "error",
            summary: "Sandbox konnte nicht erstellt werden!",
            detail: errorMessage,
            life: 6000,
          });
        }
      } catch (error) {
        const errorMessage =
          error.response?.data.message || error.message || error;
        console.error("Fehler beim Erstellen der Sandbox:", errorMessage);
        this.$toast.add({
          severity: "error",
          summary: "Sandbox konnte nicht erstellt werden!",
          detail: errorMessage,
          life: 6000,
        });
      } finally {
        const elapsedTime = Date.now() - startTime;

        const minLoadingTime = 8000; // 8 Sekunden
        const waitTime = Math.max(minLoadingTime - elapsedTime, 0);

        setTimeout(() => {
          this.generalStore.setLoading(false);
          console.log("Stop loading");
        }, waitTime);
      }
    },
  },
};
</script>

<template>
  <Card style="width: 25rem; overflow: hidden">
    <template #header>
      <img :src="image.thumbnail" alt="Shopware 6 Sandbox" />
    </template>
    <template #title>{{ image.title }}</template>
    <template #subtitle>{{ image.imageName }}</template>
    <template #footer>
      <div class="flex gap-4 mt-1">
        <Button label="Starten" class="w-full" @click="createSandbox" />
      </div>
    </template>
  </Card>
</template>

<style scoped></style>
