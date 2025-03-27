<script setup>
import Menu from "primevue/menu";
import { ref } from "vue";

const items = ref([
  {
    label: "Sandbox Environments",
    icon: "pi pi-box",
    route: "/admin/sandbox-environments",
  },
  {
    label: "Sandbox Images",
    icon: "pi pi-image",
    route: "/admin/sandbox-images",
  },
  {
    label: "Audit Log",
    icon: "pi pi-align-left",
    route: "/admin/auditlog",
  },
]);
</script>

<template>
  <Menu :model="items">
    <template #item="{ item, props }">
      <router-link v-if="item.route" :to="item.route" v-bind="props.action">
        <a v-ripple>
          <span :class="item.icon" />
          <span class="ml-2">{{ item.label }}</span>
        </a>
      </router-link>

      <a
        v-else
        v-ripple
        :href="item.url"
        :target="item.target"
        v-bind="props.action"
      >
        <span :class="item.icon" />
        <span class="ml-2">{{ item.label }}</span>
      </a>
    </template>
  </Menu>
</template>
