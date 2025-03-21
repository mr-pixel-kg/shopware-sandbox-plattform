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

  data() {
    return {
      title: "Shopware 6.6.10.0",
      subtitle: "dockware/dev:6.6.10.0",
      thumbnail: "/shopware-banner.jpg",
      image: "image",
      sandboxImage: "dockware/dev:6.6.10.0"
    };
  },

  setup() {
    const store = GeneralStore();
    return {
      generalStore: store
    }
  },

  methods: {
    async createSandbox() {

      this.generalStore.setLoading(true);
      console.log("Start loading")

      const startTime = Date.now()

      try {
        const response = await sandboxService.createSandbox(this.sandboxImage, 60)

        if (response.status === "success") {
          console.log("Sandbox erfolgreich erstellt", response);
        } else {
          const errorMessage = response.message || "Unbekannter Fehler";
          console.log("Fehler beim Erstellen der Sandbox:", errorMessage);
        }
      } catch(error) {
        console.error("Fehler beim Erstellen der Sandbox:", error.response?.data.message || error.message || error);
      } finally {
        const elapsedTime = Date.now() - startTime;

        const minLoadingTime = 8000; // 8 Sekunden
        const waitTime = Math.max(minLoadingTime - elapsedTime, 0);

        setTimeout(() => {
          this.generalStore.setLoading(false);
          console.log("Stop loading")
        }, waitTime);
      }

    }
  }

};
</script>

<template>
  <Card style="width: 25rem; overflow: hidden">
    <template #header>
      <img :src="thumbnail" alt="Shopware 6 Sandbox">
    </template>
    <template #title>{{ title }}</template>
    <template #subtitle>{{ subtitle }}</template>
    <template #footer>
      <div class="flex gap-4 mt-1">
        <Button label="Starten" class="w-full" @click="createSandbox" />
      </div>
    </template>
  </Card>
</template>

<style scoped>

</style>