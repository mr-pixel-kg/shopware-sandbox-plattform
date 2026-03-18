<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import AppHeader from "@/components/layout/AppHeader.vue";
import { Card, CardContent } from "@/components/ui/card";
import { api } from "@/lib/api";
import type { ImageRecord, SandboxRecord } from "@/types/api";
import PublicImageCard from "@/features/public/PublicImageCard.vue";
import GuestSandboxList from "@/features/public/GuestSandboxList.vue";

const images = ref<ImageRecord[]>([]);
const sandboxes = ref<SandboxRecord[]>([]);
const loading = ref(true);
const error = ref<string | null>(null);
const creatingId = ref<string | null>(null);
const deletingId = ref<string | null>(null);

const heroCount = computed(() => `${images.value.length} demo templates`);

async function load() {
  loading.value = true;
  error.value = null;

  try {
    const [publicImages, guestSandboxes] = await Promise.all([
      api.getPublicImages(),
      api.getGuestSandboxes(),
    ]);
    images.value = publicImages;
    sandboxes.value = guestSandboxes;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Could not load storefront";
  } finally {
    loading.value = false;
  }
}

async function startDemo(imageId: string) {
  creatingId.value = imageId;
  try {
    await api.createDemo({ imageId });
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Could not create demo";
  } finally {
    creatingId.value = null;
  }
}

async function deleteGuestSandbox(id: string) {
  deletingId.value = id;
  try {
    await api.deleteGuestSandbox(id);
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Could not delete demo sandbox";
  } finally {
    deletingId.value = null;
  }
}

onMounted(load);
</script>

<template>
  <div>
    <AppHeader />

    <div class="app-shell">
      <section class="mb-8">
        <div class="space-y-3">
          <h1 class="text-3xl font-semibold tracking-tight md:text-4xl">Public Images</h1>
          <p class="max-w-3xl text-sm text-muted-foreground md:text-base">
            Start a demo from the public image gallery and keep your active guest sandboxes in view on the right.
          </p>
        </div>
      </section>

      <div v-if="error" class="mb-6 rounded-2xl border border-danger/20 bg-danger/10 px-4 py-3 text-sm text-danger">
        {{ error }}
      </div>

      <div class="grid gap-8 xl:grid-cols-[minmax(0,1fr)_360px]">
        <section class="space-y-5">
          <div class="flex items-center justify-between gap-4">
            <div>
              <h2 class="section-title">Gallery</h2>
              <p class="text-sm text-muted-foreground">Published image templates available for guests.</p>
            </div>

            <Card>
              <CardContent class="px-4 py-3">
                <div class="text-sm font-semibold text-foreground">{{ heroCount }}</div>
              </CardContent>
            </Card>
          </div>

          <div v-if="loading" class="rounded-xl border border-dashed border-border bg-secondary/30 px-4 py-12 text-center text-sm text-muted-foreground">
            Loading public image catalog...
          </div>

          <div v-else class="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
            <PublicImageCard
              v-for="image in images"
              :key="image.id"
              :image="image"
              :busy="creatingId === image.id"
              @demo="startDemo"
            />
          </div>
        </section>

        <aside class="xl:sticky xl:top-6 xl:self-start">
          <GuestSandboxList :sandboxes="sandboxes" :deleting-id="deletingId" @remove="deleteGuestSandbox" />
        </aside>
      </div>
    </div>
  </div>
</template>
