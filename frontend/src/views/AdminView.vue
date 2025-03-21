<script setup>
import Sidebar from "@/components/admin/Sidebar.vue";
import Panel from "primevue/panel";
</script>

<template>
  <div class="flex h-screen">
    <div id="sidebar" class="w-1/4 p-3">
      <Sidebar></Sidebar>
    </div>
    <div class="w-3/4 p-3">
      <Panel>
        <router-view></router-view>
      </Panel>
    </div>
  </div>
</template>

<style scoped>
#sidebar {
  max-width: 320px;
}
</style>
