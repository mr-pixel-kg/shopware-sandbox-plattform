<script>
import Card from "primevue/card";
import Button from "primevue/button";
import { SandboxImageModel } from "@/models/SandboxImageModel.js";
import SandboxService from "@/services/sandboxService.js";
import { GeneralStore } from "@/stores/generalStore.js";

export default {
  name: "ActiveSandboxImageCard",

  props: {
    sandboxImage: {
      type: SandboxImageModel,
      required: true,
    },
  },

  components: {
    Card,
    Button,
  },

  setup() {
    const store = GeneralStore();
    return {
      generalStore: store,
    };
  },

  methods: {
    async createDemo() {
      console.log("Create Demo");

      this.generalStore.setLoading(true);
      console.log("Start loading");

      const startTime = Date.now();
      let sandbox = null;

      try {
        const resp = await SandboxService.createSandbox(
          this.sandboxImage.imageName,
          60,
        );

        if (resp.success) {
          sandbox = resp.sandbox;
          console.log("Sandbox created", sandbox);
        } else {
          const errorMessage = resp.message;
          console.error("Fehler beim Erstellen der Sandbox:", errorMessage);

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

        const minLoadingTime = 5000; // 8 Sekunden
        const waitTime = Math.max(minLoadingTime - elapsedTime, 0);

        setTimeout(() => {
          this.generalStore.setLoading(false);
          console.log("Stop loading");

          if(sandbox !== null) {
            this.generalStore.addSandbox(sandbox);

            this.$toast.add({
              severity: "success",
              summary: this.sandboxImage.title,
              detail: "Sandbox erfolgreich erstellt!",
              life: 3000,
            });
          }
        }, waitTime);
      }
    },
  },
};
</script>

<template>
  <Card style="overflow: hidden" class="w-full">
    <template #header>
      <img alt="Sandbox Image Thumbnail" :src="sandboxImage.thumbnailUrl" />
    </template>
    <template #title>{{ sandboxImage.title }}</template>
    <template #subtitle>{{ sandboxImage.imageName }}</template>
    <template #content>
      <p class="m-0">
        {{ sandboxImage.description }}
      </p>
    </template>
    <template #footer>
      <div class="flex gap-4 mt-1">
        <a :href="sandboxImage.infoLink" class="w-full">
          <Button
            label="Zum Store"
            severity="secondary"
            class="w-full"
            outlined
          />
        </a>
        <Button label="Demo" class="w-full" @click="createDemo" />
      </div>
    </template>
  </Card>
</template>

<style scoped></style>
