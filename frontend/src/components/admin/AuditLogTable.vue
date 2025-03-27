<script>
import DataTable from "primevue/datatable";
import Column from "primevue/column";
import Button from "primevue/button";
import Tag from "primevue/tag";
import Dialog from "primevue/dialog";
import InputText from "primevue/inputtext";
import ProgressSpinner from "primevue/progressspinner";
import AuditLogService from "@/services/auditLogService.js";
import * as UAParser from "ua-parser-js"
import {computed} from "vue";
import {Badge} from "primevue";

export default {
  components: {
    DataTable,
    Column,
    Button,
    Tag,
    Badge,
    Dialog,
    InputText,
    ProgressSpinner,
  },

  data() {
    return {
      logEntries: [],
    };
  },

  methods: {
    async loadData() {
      try {
        this.logEntries = await AuditLogService.getAllLogEntries()
      } catch (error) {
        console.error("Failed to load audit log entries:", error.message);
      }
    },
    parseUserAgent(userAgent) {
      const parser = new UAParser.UAParser(userAgent); // ✅ Verwende `UAParser.UAParser`
      const result = parser.getResult();

      const browser = computed(() => result.browser.name || "Unknown");
      const os = computed(() => result.os.name || "Unknown");
      const device = computed(() => result.device.type || "Desktop");

      return { browser: browser, os: os, device: device };
    },
    getActionSeverity(action) {
      switch (action) {
        case "SANDBOX_CREATE":
        case "USER_LOGIN":
          return "success";
        case "USER_LOGIN_FAILED":
          return "danger";
        case "IMAGE_CREATE":
          return "info";
        case "SANDBOX_DELETE":
        case "IMAGE_DELETE":
        case "CONTAINER_AUTO_REMOVE":
          return "danger";
        default:
          return "contrast";
      }
    },
  },

  mounted() {
    this.loadData();
  },
};
</script>

<template>
  <div class="card">
    <!-- Data Table -->
    <DataTable :value="logEntries" tableStyle="min-width: 50rem">
      <template #header>
        <div class="flex items-center justify-between gap-2">
          <span class="text-xl font-bold">Audit Log</span>
          <Button icon="pi pi-refresh" rounded raised @click="loadData" />
        </div>
      </template>

      <!-- Columns -->
      <Column field="timestamp" header="Timestamp"></Column>
      <Column field="ip_address" header="IP Address"></Column>
      <Column field="user_agent" header="User Agent">
        <template #body="{ data }">
          <div class="flex gap-2">
            <Badge :value="parseUserAgent(data.userAgent).browser" severity="info" />
            <Badge :value="parseUserAgent(data.userAgent).os" severity="success" />
            <Badge :value="parseUserAgent(data.userAgent).device" severity="warning" />
          </div>
        </template>
      </Column>
      <Column field="username" header="User"></Column>
      <Column field="action" header="Action">
        <template #body="{ data }">
          <Tag :value="data.action" :severity="getActionSeverity(data.action)" />
        </template>
      </Column>

      <!-- Empty State -->
      <template #empty>
        <div class="text-center text-gray-500">
          <i class="pi pi-info-circle text-xl"></i>
          <p>No data available!</p>
        </div>
      </template>
    </DataTable>
  </div>
</template>

<style scoped></style>
