<script>
import DataTable from "primevue/datatable";
import Column from "primevue/column";
import Button from "primevue/button";
import Tag from "primevue/tag";
import Select from "primevue/select";
import Dialog from "primevue/dialog";
import ProgressSpinner from "primevue/progressspinner";
import SandboxService from "@/services/sandboxService.js";
import ImagesService from "@/services/imagesService.js";

export default {
  components: {
    DataTable,
    Column,
    Button,
    Select,
    Dialog,
    Tag,
    ProgressSpinner,
  },

  data() {
    return {
      sandboxes: [],
      createSandboxDialog: {
        visible: false,
        loading: false,
        form: {
          image: "",
          lifetime: -1,
        },
      },
      availableSandboxImages: [],
      availableSandboxLifetimes: [
        { display: "1 hour", value: 60 },
        { display: "1 day", value: 1440 },
        { display: "Endless", value: -1 },
      ],
    };
  },

  methods: {
    async loadSandboxes() {
      try {
        this.sandboxes = await SandboxService.getAllSandboxes();
      } catch (error) {
        console.error("Failed to load sandboxes:", error.message);
      }
    },

    async openCreateSandboxDialog() {
      try {
        const result = await ImagesService.getAllImages();
        this.availableSandboxImages = result.map(
          (item) => `${item.image_name}:${item.image_tag}`,
        );
        this.createSandboxDialog.visible = true;
      } catch (error) {
        console.error("Failed to load images:", error.message);
      }
    },

    async createSandboxEnvironment() {
      const { image, lifetime } = this.createSandboxDialog.form;

      this.createSandboxDialog.loading = true;
      try {
        await SandboxService.createSandbox(image, lifetime);
        this.loadSandboxes();
        this.resetCreateSandboxForm();
      } catch (error) {
        console.error("Failed to create sandbox:", error.message);
      } finally {
        this.createSandboxDialog.loading = false;
        this.createSandboxDialog.visible = false;
      }
    },

    async deleteSandbox(sandbox) {
      try {
        await SandboxService.deleteSandbox(sandbox.id);
        this.loadSandboxes();
      } catch (error) {
        console.error("Failed to delete sandbox:", error.message);
      }
    },

    openSandboxWindow(sandbox) {
      window.open(`https://${sandbox.url}`, "_blank").focus();
    },

    getStatus(sandbox) {
      return sandbox.state === "running" ? "success" : "danger";
    },

    getFormattedDateTime(datetime) {
      if (!datetime) return "N/A";
      const date = new Date(datetime);
      return date.toLocaleString();
    },

    getRemainingTime(destroyAt) {
      if (!destroyAt) return "never";
      const now = new Date();
      const destroyDate = new Date(destroyAt);
      const diffMs = destroyDate - now;
      const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
      const diffHrs = Math.floor(
        (diffMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60),
      );
      const diffMins = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60));

      if (diffDays > 0) {
        return `${diffDays} days`;
      } else if (diffHrs > 0) {
        return `${diffHrs}:${diffMins}`;
      } else {
        return `00:${diffMins}`;
      }
    },

    resetCreateSandboxForm() {
      this.createSandboxDialog.form = {
        image: "",
        lifetime: -1,
      };
    },
  },

  mounted() {
    this.loadSandboxes();
  },
};
</script>

<template>
  <div class="card">
    <!-- DataTable -->
    <DataTable :value="sandboxes" tableStyle="min-width: 50rem">
      <template #header>
        <div class="flex items-center justify-between gap-2">
          <span class="text-xl font-bold">Sandbox Environments</span>
          <div class="flex gap-2">
            <Button
              icon="pi pi-plus"
              rounded
              raised
              @click="openCreateSandboxDialog"
            />
            <Button
              icon="pi pi-refresh"
              rounded
              raised
              @click="loadSandboxes"
            />
          </div>
        </div>
      </template>

      <!-- Columns -->
      <Column field="id" header="ID"></Column>
      <Column field="image" header="Image"></Column>
      <Column field="created_at" header="Created">
        <template #body="{ data }">
          {{ getFormattedDateTime(data.created_at) }}
        </template>
      </Column>
      <Column field="destroy_at" header="Destroy">
        <template #body="{ data }">
          <Tag
            :value="getRemainingTime(data.destroy_at)"
            severity="secondary"
          />
        </template>
      </Column>
      <Column field="state" header="Status">
        <template #body="{ data }">
          <Tag :value="data.state" :severity="getStatus(data)" />
        </template>
      </Column>
      <Column class="w-24 !text-end">
        <template #body="{ data }">
          <div class="flex gap-1 justify-center">
            <Button
              icon="pi pi-arrow-right"
              severity="secondary"
              rounded
              @click="openSandboxWindow(data)"
            />
            <Button
              icon="pi pi-trash"
              severity="secondary"
              rounded
              @click="deleteSandbox(data)"
            />
          </div>
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

  <!-- Create Sandbox Dialog -->
  <Dialog
    v-model:visible="createSandboxDialog.visible"
    modal
    header="Create Sandbox"
    :style="{ width: '25rem' }"
  >
    <div class="flex items-center gap-4 mb-4">
      <label for="sandbox-image" class="font-semibold w-24">Image</label>
      <Select
        id="sandbox-image"
        v-model="createSandboxDialog.form.image"
        :options="availableSandboxImages"
        class="flex-auto"
      />
    </div>
    <div class="flex items-center gap-4 mb-8">
      <label for="sandbox-lifetime" class="font-semibold w-24">Lifetime</label>
      <Select
        id="sandbox-lifetime"
        v-model="createSandboxDialog.form.lifetime"
        :options="availableSandboxLifetimes"
        optionLabel="display"
        optionValue="value"
        class="flex-auto"
      />
    </div>
    <div
      :class="[
        'flex gap-2',
        createSandboxDialog.loading ? 'justify-between' : 'justify-end',
      ]"
    >
      <span
        v-if="createSandboxDialog.loading"
        class="flex items-center gap-2 text-primary"
      >
        Deploying Sandbox...
        <ProgressSpinner
          style="width: 50px; height: 30px"
          strokeWidth="8"
          fill="transparent"
          animationDuration=".5s"
        />
      </span>
      <div class="flex gap-2">
        <Button
          type="button"
          label="Cancel"
          severity="secondary"
          @click="createSandboxDialog.visible = false"
        />
        <Button
          type="button"
          label="Deploy"
          :disabled="createSandboxDialog.loading"
          @click="createSandboxEnvironment"
        />
      </div>
    </div>
  </Dialog>
</template>

<style scoped></style>
