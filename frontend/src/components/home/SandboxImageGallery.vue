<script>
import Card from "primevue/card";
import Button from "primevue/button";
import ActiveSandboxImageCard from "@/components/home/SandboxImageCard.vue";
import ActiveSandboxCard from "@/components/home/SandboxCard.vue";
import { SandboxImageModel } from "@/models/SandboxImageModel.js";

export default {
  name: "SandboxImageGallery",

  components: {
    ActiveSandboxCard,
    ActiveSandboxImageCard,
    Card,
    Button,
  },

  data() {
    return {
      sandboxImages: [
        new SandboxImageModel(
          "Shopware 6.6",
          "shopware-banner.jpg",
          "mr-pixel/sw-sandbox:6.6.10.0",
          "Eine vollständige Shopware 6.6.10.0 Entwicklungsumgebung, ausgestattet mit allen notwendigen Dev-Tools. Ideal für Entwickler.",
          "https://www.mr-pixel.de",
        )
        /*new SandboxImageModel(
          "Shopware 6.6 Play",
          "shopware-banner.jpg",
          "dockware/play:6.6.10.0",
          "Eine sofort einsatzbereite, saubere Shopware 6.6.10.0 Installation zum Testen, Experimentieren und Erkunden.",
          "https://www.mr-pixel.de",
        ),*/
      ],
    };
  },
};
</script>

<template>
  <div class="max-w-7xl mx-auto p-5">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-200">Shopware Umgebungen</h1>
    </div>
    <div
      class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6"
      id="theme-grid"
    >
      <ActiveSandboxImageCard
        v-for="sandboxImage in sandboxImages"
        :sandboxImage="sandboxImage"
      />
    </div>
  </div>
</template>

<style scoped></style>
