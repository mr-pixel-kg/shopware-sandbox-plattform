<script>
import Button from "primevue/button";
import Select from "primevue/select";
import SandboxImageCard from "@/components/home/SandboxImageCard.vue";
import SandboxCard from "@/components/home/SandboxCard.vue";
import LoadingScreen from "@/components/home/LoadingScreen.vue";
import { GeneralStore } from "@/stores/generalStore.js";
import { SandboxEnvironment } from "@/models/SandboxEnvironment.js";
import { SandboxImage } from "@/models/SandboxImage.js";
import SandboxImageGallery from "@/components/home/SandboxImageGallery.vue";
import SandboxGallery from "@/components/home/SandboxGallery.vue";

export default {
  // Properties returned from data() become reactive state
  // and will be exposed on `this`.
  components: {
    SandboxGallery,
    SandboxImageGallery,
    SandboxImageCard,
    LoadingScreen,
    SandboxCard,
    Button,
    Select,
  },

  data() {
    return {
      /*environments: [
        new SandboxEnvironment("uuid-8r94rhiweofnadsifuhsdifudsif","shopware/image:latest", "myUrl"),
        new SandboxEnvironment("uuid-8r94rhiweofnadsifuhsdifudsif","shopware/image:latest", "myUrl"),
        new SandboxEnvironment("uuid-8r94rhiweofnadsifuhsdifudsif","shopware/image:latest", "myUrl"),
      ],*/
      sandboxImages: [
        new SandboxImage(
          "Shopware 6.6",
          "dockware/dev:6.6.10.0",
          "/shopware-banner.jpg",
        ),
        new SandboxImage(
          "Shopware Play",
          "dockware/dev:6.6.10.0",
          "/shopware-banner.jpg",
        ),
      ],
      /*shopwareVersion: "v6.6.0.0",
      versions: [
        { name: 'v6.6', code: 'v6.6.0.0' },
        { name: 'v6.5', code: 'v6.6.5.0' },
        { name: 'v6.4', code: 'v6.6.4.0' },
      ],
      plugin: null,
      plugins: [
        { name: 'MrpixCloudPrint', repo: 'todo' },
        { name: 'MrpixGastronomy', repo: 'todo' },
        { name: 'MrpixColorTabs', repo: 'todo' },
      ]*/
    };
  },

  computed: {
    environments() {
      return this.generalStore.getSandboxEnvironments;
    },
  },

  setup() {
    const store = GeneralStore();
    return {
      generalStore: store,
    };
  },
};
</script>

<template>

  <div class="max-w-7xl mx-auto">
    <SandboxGallery />
  </div>

  <div class="max-w-7xl mx-auto">
    <SandboxImageGallery />
  </div>

  <div class="max-w-7xl mx-auto mt-10 ml-10 mr-10">
    <div v-if="environments.length">
      <div class="flex justify-between items-center mb-6">
        <h1 class="text-2xl font-bold text-gray-900">Aktive Sandboxes</h1>
      </div>
      <div
        class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6"
        id="theme-grid"
      >
        <template v-for="sandbox in this.environments">
          <SandboxCard :sandboxEnvironment="sandbox"></SandboxCard>
        </template>
      </div>
    </div>

    <div class="flex justify-between items-center mb-6 mt-8">
      <h1 class="text-2xl font-bold text-gray-900">Sandbox Umgebungen</h1>
    </div>
    <div
      class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6"
      id="theme-grid"
    >
      <template v-for="image in this.sandboxImages">
        <SandboxImageCard :sandboxImage="image"></SandboxImageCard>
      </template>
    </div>
  </div>

  <LoadingScreen></LoadingScreen>
</template>

<style scoped></style>
