<script>
import DataTable from "primevue/datatable";
import Column from "primevue/column";
import Button from "primevue/button";
import Tag from "primevue/tag";
import Dialog from "primevue/dialog";
import InputText from "primevue/inputtext";
import ProgressSpinner from "primevue/progressspinner";
import ImagesService from "@/services/imagesService.js";

export default {
  components: {
    DataTable,
    Column,
    Button,
    Tag,
    Dialog,
    InputText,
    ProgressSpinner,
  },

  data() {
    return {
      images: [],
      addImageDialogVisible: false,
      pullingImageLoading: false,
      imageForm: {
        name: "",
        tag: "",
      },
    };
  },

  methods: {
    async loadData() {
      try {
        this.images = await ImagesService.getAllImages();
      } catch (error) {
        console.error("Failed to load images:", error.message);
      }
    },

    formatSize(bytes) {
      const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
      if (bytes === 0) return "0 Bytes";
      const i = Math.floor(Math.log(bytes) / Math.log(1024));
      const size = bytes / Math.pow(1024, i);
      return `${size.toFixed(2)} ${sizes[i]}`;
    },

    async deleteImage(data) {
      try {
        await ImagesService.deleteImage(data.id);
        await this.loadData();
      } catch (error) {
        console.error("Failed to delete image:", error.message);
      }
    },

    async addSandboxImage() {
      this.pullingImageLoading = true;

      try {
        await ImagesService.registerImage(
          this.imageForm.name,
          this.imageForm.tag,
        );
        await this.loadData();
        this.resetImageForm();
        this.addImageDialogVisible = false;
      } catch (error) {
        console.error("Failed to add sandbox image:", error.message);
      } finally {
        this.pullingImageLoading = false;
      }
    },

    resetImageForm() {
      this.imageForm = {
        name: "",
        tag: "",
      };
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
    <DataTable :value="images" tableStyle="min-width: 50rem">
      <template #header>
        <div class="flex items-center justify-between gap-2">
          <span class="text-xl font-bold">Sandbox Images</span>
          <div class="flex gap-2">
            <Button
              icon="pi pi-plus"
              rounded
              raised
              @click="addImageDialogVisible = true"
            />
            <Button icon="pi pi-refresh" rounded raised @click="loadData" />
          </div>
        </div>
      </template>

      <!-- Columns -->
      <Column field="id" header="ID">
        <template #body="{ data }"> {{ data.id.slice(0, 10) }}... </template>
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
          <div class="flex gap-1 justify-center">
            <Button
              icon="pi pi-trash"
              severity="secondary"
              rounded
              @click="deleteImage(data)"
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

  <!-- Dialog -->
  <Dialog
    v-model:visible="addImageDialogVisible"
    modal
    header="Add Sandbox Image"
    :style="{ width: '25rem' }"
  >
    <span class="text-surface-500 dark:text-surface-400 block mb-8">
      Enter docker image:
    </span>
    <div class="flex items-center gap-4 mb-4">
      <label for="image-name" class="font-semibold w-24">Image Name</label>
      <InputText
        id="image-name"
        v-model="imageForm.name"
        class="flex-auto"
        autocomplete="off"
      />
    </div>
    <div class="flex items-center gap-4 mb-8">
      <label for="image-tag" class="font-semibold w-24">Image Tag</label>
      <InputText
        id="image-tag"
        v-model="imageForm.tag"
        class="flex-auto"
        autocomplete="off"
      />
    </div>
    <div
      :class="[
        'flex gap-2',
        pullingImageLoading ? 'justify-between' : 'justify-end',
      ]"
    >
      <span
        v-if="pullingImageLoading"
        class="flex items-center gap-2 text-primary"
      >
        Pulling image...
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
          @click="addImageDialogVisible = false"
        />
        <Button
          type="button"
          label="Save"
          :disabled="pullingImageLoading"
          @click="addSandboxImage"
        />
      </div>
    </div>
  </Dialog>
</template>

<style scoped></style>
