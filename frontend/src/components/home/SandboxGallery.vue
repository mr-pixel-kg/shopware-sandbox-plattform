<script>
import ActiveSandboxCard from "@/components/home/SandboxCard.vue";
import { GeneralStore } from "@/stores/generalStore.js";

export default {
  name: "SandboxGallery",

  components: { ActiveSandboxCard },

  methods: {
    deleteSandbox(id) {
      // Is this needed?
      this.generalStore.removeSandbox(id);
    },
  },

  setup() {
    const store = GeneralStore();
    return {
      generalStore: store,
    };
  },

  mounted() {
    setInterval(() => {
      this.generalStore.refreshSandboxes();
    }, 5000);
  }
};
</script>

<template>
  <div class="max-w-7xl mx-auto p-5">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-200">Meine Sandbox Umgebungen</h1>
    </div>

    <div
      class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6"
      id="theme-grid"
    >
      <ActiveSandboxCard
        v-for="sandbox in generalStore.sandboxes"
        :key="sandbox.id"
        :sandbox="sandbox"
        @delete-sandbox="deleteSandbox"
      />
    </div>
  </div>
</template>

<style scoped></style>
