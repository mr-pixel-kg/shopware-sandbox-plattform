<script>
import Card from "primevue/card";
import Button from "primevue/button";
import Tag from "primevue/tag";
import { ProgressBar } from "primevue";
import { SandboxEnvironmentModel } from "@/models/SandboxEnvironmentModel.js";
import SandboxService from "@/services/sandboxService.js";
import { GeneralStore } from "@/stores/generalStore.js";

export default {
  name: "ActiveSandboxCard",

  components: {
    Card,
    Button,
    Tag,
    ProgressBar,
  },

  props: {
    sandbox: {
      type: SandboxEnvironmentModel,
      required: true,
    },
  },

  data() {
    return {
      refreshInterval: null,
      remainingTime: "0m 1s",
    };
  },

  computed: {
    statusColor() {
      // This does not work to refresh tag color when status changes
      switch (this.sandbox.status) {
        case "running":
          return "bg-green-500";
        case "stopped":
          return "bg-yellow-500";
        case "created":
          return "bg-yellow-500";
        case "restarting":
          return "bg-yellow-500";
        case "exited":
          return "bg-red-500";
        case "dead":
          return "bg-red-500";
        case "paused":
          return "bg-gray-500";
        default:
          return "bg-gray-500";
      }
    },
  },

  setup() {
    const store = GeneralStore();
    return {
      generalStore: store,
    };
  },

  mounted() {
    this.refreshData();

    this.refreshInterval = setInterval(() => {
      this.refreshData();
    }, 1000);
  },

  beforeUnmount() {
    clearInterval(this.refreshInterval);
  },

  methods: {
    openUrl(url) {
      window.open(url, "_blank");
    },

    async deleteSandbox() {
      const resp = await SandboxService.deleteSandbox(this.sandbox.id);

      if (resp.success === true) {
        this.$toast.add({
          severity: "success",
          summary: "Sandbox gelöscht",
          detail: resp.message,
          life: 3000,
        });
        this.generalStore.removeSandbox(this.sandbox.id);
        this.$emit("delete-sandbox", this.sandbox.id); // is this needed?
      } else {
        this.$toast.add({
          severity: "error",
          summary: "Sandbox löschen fehlgeschlagen",
          detail: resp.message,
          life: 6000,
        });
      }
    },

    refreshData() {
      this.remainingTime = this.sandbox.getRemainingTime();
    },
  },
};
</script>

<template>
  <Card
    class="w-full max-w-lg shadow-lg border border-gray-200"
    style="overflow: hidden"
  >
    <template #title>Shopware Sandbox</template>
    <template #header>
      <ProgressBar
        :value="100 - (this.remainingTime.split('m')[0] * 100) / 60"
        :show-value="false"
        style="height: 5px"
      ></ProgressBar>
    </template>
    <template #subtitle>
      <span class="text-gray-500">{{ sandbox.imageName }}</span>
    </template>

    <template #content>
      <div class="space-y-2">
        <div class="flex justify-between items-center">
          <span class="font-semibold">Status:</span>
          <Tag :class="[statusColor, 'px-2 py-1 text-white']">{{
            sandbox.status
          }}</Tag>
        </div>
        <div class="flex justify-between items-center">
          <span class="font-semibold">Läuft bis:</span>
          <span class="text-gray-700">{{ this.remainingTime }}</span>
        </div>

        <div class="bg-gray-100 p-3 rounded-lg mt-6">
          <h3 class="font-semibold text-gray-600 mb-2">Zugangsdaten:</h3>
          <div class="flex justify-between">
            <span class="">Benutzername:</span>
            <span class="text-gray-800">admin</span>
          </div>
          <div class="flex justify-between">
            <span class="">Passwort:</span>
            <span class="text-gray-800">shopware</span>
          </div>
        </div>
      </div>
    </template>

    <template #footer>
      <div class="flex gap-2 mt-3">
        <Button
          icon="pi pi-trash"
          severity="danger"
          class="w-1/4"
          @click="deleteSandbox"
        />
        <a class="w-1/2" :href="sandbox.getStorefrontUrl()" target="_blank">
          <Button
            label="Storefront"
            severity="primary"
            class="w-full"
            @click="openUrl(sandbox.getStorefrontUrl())"
          />
        </a>

        <a class="w-1/2" :href="sandbox.getAdminUrl()" target="_blank">
          <Button
            label="Admin"
            severity="secondary"
            class="w-full"
            @click="openUrl(sandbox.getAdminUrl())"
          />
        </a>
      </div>
    </template>
  </Card>
</template>

<style scoped></style>
