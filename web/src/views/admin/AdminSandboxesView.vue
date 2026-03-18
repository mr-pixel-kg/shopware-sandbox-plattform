<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import SandboxesManager from "@/features/admin/SandboxesManager.vue";
import { api } from "@/lib/api";
import { useAuthStore } from "@/stores/auth";
import type { CreateImagePayload, CreateSandboxPayload, ImageRecord, SandboxRecord } from "@/types/api";

const auth = useAuthStore();

const images = ref<ImageRecord[]>([]);
const sandboxes = ref<SandboxRecord[]>([]);
const loading = ref(true);
const error = ref<string | null>(null);
const creatingSandbox = ref(false);
const deletingSandboxId = ref<string | null>(null);
const snapshottingSandboxId = ref<string | null>(null);

const token = computed(() => auth.token);

async function load() {
  if (!token.value) return;

  loading.value = true;
  error.value = null;

  try {
    const [imageResult, sandboxResult] = await Promise.all([
      api.getImages(token.value),
      api.getSandboxes(token.value),
    ]);

    images.value = imageResult;
    sandboxes.value = sandboxResult;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Could not load sandboxes";
  } finally {
    loading.value = false;
  }
}

async function createSandbox(payload: CreateSandboxPayload) {
  if (!token.value) return;

  creatingSandbox.value = true;
  error.value = null;

  try {
    await api.createSandbox(token.value, payload);
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Could not create sandbox";
  } finally {
    creatingSandbox.value = false;
  }
}

async function deleteSandbox(id: string) {
  if (!token.value) return;

  deletingSandboxId.value = id;
  error.value = null;

  try {
    await api.deleteSandbox(token.value, id);
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Could not delete sandbox";
  } finally {
    deletingSandboxId.value = null;
  }
}

async function snapshotSandbox(id: string, payload: CreateImagePayload) {
  if (!token.value) return;

  snapshottingSandboxId.value = id;
  error.value = null;

  try {
    await api.snapshotSandbox(token.value, id, payload);
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Could not create snapshot";
  } finally {
    snapshottingSandboxId.value = null;
  }
}

onMounted(load);
</script>

<template>
  <div class="space-y-6">
    <div class="space-y-1">
      <h1 class="text-2xl font-semibold tracking-tight">Sandboxes</h1>
      <p class="text-sm text-muted-foreground">Start, inspect and snapshot running sandbox containers.</p>
    </div>

    <div v-if="error" class="rounded-md border border-destructive/20 bg-destructive/10 px-4 py-3 text-sm text-destructive">
      {{ error }}
    </div>

    <div v-if="loading" class="rounded-lg border bg-background px-4 py-8 text-sm text-muted-foreground">
      Loading sandboxes...
    </div>

    <SandboxesManager
      v-else
      :images="images"
      :sandboxes="sandboxes"
      :creating="creatingSandbox"
      :deleting-id="deletingSandboxId"
      :snapshotting-id="snapshottingSandboxId"
      @create="createSandbox"
      @remove="deleteSandbox"
      @snapshot="snapshotSandbox"
    />
  </div>
</template>
