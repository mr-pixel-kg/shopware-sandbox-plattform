<script>
import DataTable from 'primevue/datatable';
import Column from 'primevue/column';
import Button from "primevue/button";
import Tag from "primevue/tag"
import Dialog from 'primevue/dialog';
import InputText from 'primevue/inputtext';
import ProgressSpinner from 'primevue/progressspinner';
import SandboxService from "../../services/sandboxService.js";
import ImagesService from "../../services/imagesService.js";

export default {

  components: {
    DataTable,
    Column,
    Button,
    Tag,
    Dialog,
    InputText,
    ProgressSpinner
  },

  // Properties returned from data() become reactive state
  // and will be exposed on `this`.
  data() {
    return {
      images: [{
        "id": "a407dee395ed97ead1e40c7537395d6271c07cc89c317f8eda1c19f6fc783695",
        "image_name": "dockware/dev",
        "image_tag": "6.6.8.2",
        "created_at": "2024-11-12T17:10:49+01:00",
        "size": 4860039980
      }],
      addImageDialogVisible: false,
      pullingImageLoading: false,
      imageName: "",
      imageTag: ""
    }
  },

  // Methods are functions that mutate state and trigger updates.
  // They can be bound as event handlers in templates.
  methods: {
    async loadData() {
      this.images = await ImagesService.getAllImages();
    },
    formatSize(bytes) {
      const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
      if (bytes === 0) return "0 Bytes";
      const i = Math.floor(Math.log(bytes) / Math.log(1024));
      const size = bytes / Math.pow(1024, i);
      return `${size.toFixed(2)} ${sizes[i]}`;
    },
    async deleteImage(data) {
      await ImagesService.deleteImage(data.id);
      await this.loadData();
    },
    async addSandboxImage() {
      this.pullingImageLoading = true;
      await ImagesService.registerImage(this.imageName, this.imageTag);
      await this.loadData();
      this.pullingImageLoading = false;
      this.addImageDialogVisible = false;
    }
  },

  // Lifecycle hooks are called at different stages
  // of a component's lifecycle.
  // This function will be called when the component is mounted.
  mounted() {
    console.log(`Loading images table`)
    this.loadData();
  }
}
</script>

<template>
  <div class="card">
    <DataTable :value="this.images" tableStyle="min-width: 50rem">
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-2">
          <span class="text-xl font-bold">Sandbox Images</span>
          <div class="flex gap-2">
            <Button icon="pi pi-plus" rounded raised @click="addImageDialogVisible = true"/>
            <Button icon="pi pi-refresh" rounded raised @click="loadData"/>
          </div>
        </div>
      </template>
      <Column field="id" header="ID">
        <template #body="{ data }">
          {{ data.id.slice(0, 10) }}...
        </template>
      </Column>
      <Column field="image_name" header="Image Name"></Column>
      <Column field="image_tag" header="Image Tag"></Column>
      <Column field="created_at" header="Created"></Column>
      <Column field="size" header="Size">
        <template #body="{ data }">
          {{ formatSize(data.size) }}
        </template>
      </Column>
      <Column class="w-24 !text-end">
        <template #body="{ data }">
          <div class="flex /*flex-wrap*/ gap-1 justify-center">
            <Button icon="pi pi-trash" @click="deleteImage(data)" severity="secondary" rounded></Button>
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

  <Dialog v-model:visible="addImageDialogVisible" modal header="Add Sandbox Image" :style="{ width: '25rem' }">
    <span class="text-surface-500 dark:text-surface-400 block mb-8">Enter docker image:</span>
    <div class="flex items-center gap-4 mb-4">
      <label for="image-name" class="font-semibold w-24">Image Name</label>
      <InputText id="image-name" v-model="imageName" class="flex-auto" autocomplete="off" />
    </div>
    <div class="flex items-center gap-4 mb-8">
      <label for="image-tag" class="font-semibold w-24">Image Tag</label>
      <InputText id="image-tag" v-model="imageTag" class="flex-auto" autocomplete="off" />
    </div>
    <div :class="['flex gap-2', pullingImageLoading ? 'justify-between' : 'justify-end']">
      <span v-if="pullingImageLoading" class="flex items-center gap-2 text-primary">
        Pulling image...
        <ProgressSpinner style="width: 50px; height: 30px" strokeWidth="8" fill="transparent" animationDuration=".5s" />
      </span>
      <div class="flex gap-2">
        <Button type="button" label="Cancel" severity="secondary" @click="addImageDialogVisible = false"></Button>
        <Button type="button" label="Save" :disabled="pullingImageLoading" @click="addSandboxImage"></Button>
      </div>
    </div>
  </Dialog>
</template>

<style scoped>

</style>