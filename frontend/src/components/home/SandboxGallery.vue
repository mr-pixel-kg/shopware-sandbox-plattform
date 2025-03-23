<script>


import ActiveSandboxCard from "@/components/home/ActiveSandboxCard.vue";
import {SandboxEnvironmentModel} from "@/models/SandboxEnvironmentModel.js";
import { GeneralStore } from "@/stores/generalStore.js";

export default {
  name: "SandboxGallery",

  components: { ActiveSandboxCard },

  data() {
    return {
      /*sandboxes: [
        new SandboxEnvironmentModel(
            "3ba79559-6383-4848-80f7-b38aa6113ae8",
            "dockware/dev:6.6.8.2",
            "running",
            Date.now(),
            new Date(Date.now() + 60 * 60 * 1000), // Ablaufzeit in 60 Min
            "https://sandbox-0a185659-d844-4e59-898f-5ddb857edfde.shopshredder.de",
        )
      ]*/
    };
  },

  methods: {
    deleteSandbox(id) {
      console.log("This is a test if this method is called");
      this.generalStore.removeSandbox2(id);
      //this.sandboxes = this.sandboxes.filter(sandbox => sandbox.id !== id);
    }
  },

  setup() {
    const store = GeneralStore();
    return {
      generalStore: store,
    };
  },
}
</script>

<template>
  <div class="max-w-7xl mx-auto bg-red-500 p-5">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-gray-900">Meine Sandboxen</h1>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6" id="theme-grid">

      <ActiveSandboxCard
          v-for="sandbox in generalStore.sandboxes"
          :key="sandbox.id"
          :sandbox="sandbox"
          @delete-sandbox="deleteSandbox"
      />

    </div>

  </div>
</template>

<style scoped>

</style>