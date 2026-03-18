<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import ImagesManager from "@/features/admin/ImagesManager.vue";
import { api } from "@/lib/api";
import { useAuthStore } from "@/stores/auth";
import type { CreateImagePayload, ImageRecord } from "@/types/api";

const auth = useAuthStore();

const images = ref<ImageRecord[]>([]);
const loading = ref(true);
const error = ref<string | null>(null);
const creatingImage = ref(false);
const deletingImageId = ref<string | null>(null);

const token = computed(() => auth.token);

async function load() {
  if (!token.value) return;

  loading.value = true;
  error.value = null;

  try {
    images.value = await api.getImages(token.value);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Could not load images";
  } finally {
    loading.value = false;
  }
}

async function createImage(payload: CreateImagePayload) {
  if (!token.value) return;

  creatingImage.value = true;
  error.value = null;

  try {
    await api.createImage(token.value, payload);
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Could not create image";
  } finally {
    creatingImage.value = false;
  }
}

async function deleteImage(id: string) {
  if (!token.value) return;

  deletingImageId.value = id;
  error.value = null;

  try {
    await api.deleteImage(token.value, id);
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Could not delete image";
  } finally {
    deletingImageId.value = null;
  }
}

onMounted(load);
</script>

<template>
  <div class="space-y-6">
    <div class="space-y-1">
      <h1 class="text-2xl font-semibold tracking-tight">Images</h1>
      <p class="text-sm text-muted-foreground">Register and manage Docker image templates.</p>
    </div>

    <div v-if="error" class="rounded-md border border-destructive/20 bg-destructive/10 px-4 py-3 text-sm text-destructive">
      {{ error }}
    </div>

    <div v-if="loading" class="rounded-lg border bg-background px-4 py-8 text-sm text-muted-foreground">
      Loading images...
    </div>

    <ImagesManager
      v-else
      :images="images"
      :creating="creatingImage"
      :deleting-id="deletingImageId"
      @create="createImage"
      @remove="deleteImage"
    />
  </div>
</template>
