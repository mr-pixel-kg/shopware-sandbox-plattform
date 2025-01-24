<script>
import DataTable from 'primevue/datatable';
import Column from 'primevue/column';
import Button from "primevue/button";
import Tag from "primevue/tag";
import Select from 'primevue/select';
import Dialog from 'primevue/dialog';
import ProgressSpinner from 'primevue/progressspinner';
import SandboxService from "../../services/sandboxService.js";
import ImagesService from "../../services/imagesService.js";

export default {

  components: {
    DataTable,
    Column,
    Button,
    Select,
    Dialog,
    Tag,
    ProgressSpinner
  },

  // Properties returned from data() become reactive state
  // and will be exposed on `this`.
  data() {
    return {
      sandboxes: [{
        id: "aafb929f-0992-44d2-a8d3-17619360deff",
        container_id: "1fdab6cd2c18d260586b462ec7cd86f482d5b7fef6ee84fdb3733b97b79ae652",
        container_name: "/sandbox-aafb929f-0992-44d2-a8d3-17619360deff",
        url: "sandbox-aafb929f-0992-44d2-a8d3-17619360deff.shopshredder.zion.mr-pixel.de",
        image: "dockware/dev:6.6.8.2",
        created_at: "2024-12-20T13:08:50+01:00",
        state: "running",
        status: "Up 1 second"
      }],
      createSandboxDialogVisible: false,
      deployingSandboxLoading: false,
      sandboxImage: "",
      availableSandboxImages: ["dockware/dev:6.6.8.2"],
      sandboxLifetime: -1,
      availableSandboxLifetimes: [
        { display: 'Endless', value: -1 },
        { display: '1 hour', value: 60 },
        { display: '1 day', value: 1440 }
      ]
    }
  },

  // Methods are functions that mutate state and trigger updates.
  // They can be bound as event handlers in templates.
  methods: {
    async loadData() {
      this.sandboxes = await SandboxService.getAllSandboxes();
    },
    getStatus(sandbox) {
      switch (sandbox.state) {
        case 'running':
          return 'success';
        default:
          return "danger";
      }
    },
    openSandboxWindow(data) {
      window.open("https://" + data.url, '_blank').focus();
    },
    async deleteSandbox(data) {
      await SandboxService.deleteSandbox(data.id);
      await this.loadData();
    },
    async openCreateSandboxDialog() {
      let result = await ImagesService.getAllImages();
      this.availableSandboxImages = result.map(item => `${item.image_name}:${item.image_tag}`);
      this.createSandboxDialogVisible = true;
    },
    async createSandboxEnvironment() {
      this.deployingSandboxLoading = true;
      await SandboxService.createSandbox(this.sandboxImage, this.sandboxLifetime)
      await this.loadData();
      this.deployingSandboxLoading = false;
      this.createSandboxDialogVisible = false;
    }
  },

  // Lifecycle hooks are called at different stages
  // of a component's lifecycle.
  // This function will be called when the component is mounted.
  mounted() {
    console.log(`Loading sandbox table`)
    this.loadData();
  }
}
</script>

<template>
  <div class="card">
    <DataTable :value="sandboxes" tableStyle="min-width: 50rem">
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-2">
          <span class="text-xl font-bold">Sandbox Environments</span>
          <div class="flex gap-2">
            <Button icon="pi pi-plus" rounded raised @click="openCreateSandboxDialog"/>
            <Button icon="pi pi-refresh" rounded raised @click="loadData"/>
          </div>
        </div>
      </template>
      <Column field="id" header="ID"></Column>
      <Column field="image" header="Image"></Column>
      <Column field="created_at" header="Created"></Column>
      <Column field="state" header="Status">
        <template #body="slotProps">
          <Tag :value="slotProps.data.state" :severity="getStatus(slotProps.data)" />
        </template>
      </Column>
      <Column class="w-24 !text-end">
        <template #body="{ data }">
          <div class="flex /*flex-wrap*/ gap-1 justify-center">
            <Button icon="pi pi-arrow-right" @click="openSandboxWindow(data)" severity="secondary" rounded></Button>
            <Button icon="pi pi-trash" @click="deleteSandbox(data)" severity="secondary" rounded></Button>
          </div>
        </template>
      </Column>
      <template #empty>
        <div class="text-center text-gray-500">
          <i class="pi pi-info-circle text-xl"></i>
          <p>No data available!</p>
        </div>
      </template>
    </DataTable>
  </div>

  <Dialog v-model:visible="createSandboxDialogVisible" modal header="Create Sandbox" :style="{ width: '25rem' }">
    <div class="flex items-center gap-4 mb-4">
      <label for="sandbox-image" class="font-semibold w-24">Image</label>
      <Select id="sandbox-image" v-model="sandboxImage" :options="availableSandboxImages" class="flex-auto" autocomplete="off" />
    </div>
    <div class="flex items-center gap-4 mb-8">
      <label for="sandbox-lifetime" class="font-semibold w-24">Lifetime</label>
      <Select id="sandbox-lifetime" v-model="sandboxLifetime" :options="availableSandboxLifetimes" optionLabel="display" optionValue="value" class="flex-auto" autocomplete="off" />
    </div>
    <div :class="['flex gap-2', deployingSandboxLoading ? 'justify-between' : 'justify-end']">
      <span v-if="deployingSandboxLoading" class="flex items-center gap-2 text-primary">
        Deploying Sandbox...
        <ProgressSpinner style="width: 50px; height: 30px" strokeWidth="8" fill="transparent" animationDuration=".5s" />
      </span>
      <div class="flex gap-2">
        <Button type="button" label="Cancel" severity="secondary" @click="createSandboxDialogVisible = false"></Button>
        <Button type="button" label="Deploy" :disabled="deployingSandboxLoading" @click="createSandboxEnvironment"></Button>
      </div>
    </div>
  </Dialog>
</template>

<style scoped>

</style>