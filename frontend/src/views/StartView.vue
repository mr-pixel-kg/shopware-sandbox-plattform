<script>
import Button from 'primevue/button';
import Select from 'primevue/select';
import SandboxCard from "@/components/home/SandboxCard.vue";
import LoadingScreen from "@/components/home/LoadingScreen.vue";

export default {
  // Properties returned from data() become reactive state
  // and will be exposed on `this`.
  components: {
    LoadingScreen,
    SandboxCard,
    Button,
    Select
  },

  data() {
    return {
      shopwareVersion: "v6.6.0.0",
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
      ]
    }
  }
}
</script>

<template>

  <div class="max-w-7xl mx-auto mt-10">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-gray-900">Sandbox Umgebungen</h1>
    </div>
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6" id="theme-grid">
      <SandboxCard></SandboxCard>
      <SandboxCard></SandboxCard>
    </div>
  </div>

  <LoadingScreen></LoadingScreen>

</template>

<style scoped>

</style>