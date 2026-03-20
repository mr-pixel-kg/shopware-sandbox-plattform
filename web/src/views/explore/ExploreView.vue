<script setup lang="ts">
import { ref, computed } from 'vue'
import { useImages } from '@/composables/useImages'
import { useSandboxesStore } from '@/stores/sandboxes.store'
import { useAuthStore } from '@/stores/auth.store'
import { getApiErrorMessage } from '@/utils/error'
import { toast } from 'vue-sonner'
import PageHeader from '@/components/shared/PageHeader.vue'
import FilterBar from '@/components/explore/FilterBar.vue'
import PresetGrid from '@/components/explore/PresetGrid.vue'
import CardGridSkeleton from '@/components/shared/CardGridSkeleton.vue'
import NewSandboxDialog from '@/components/modals/NewSandboxDialog.vue'

const { images, loading } = useImages()
const sandboxesStore = useSandboxesStore()
const authStore = useAuthStore()

const filter = ref('all')
const showNewSandbox = ref(false)
const preselectedImageId = ref<string | undefined>()

const filteredImages = computed(() => {
  if (filter.value === 'all') return images.value
  if (filter.value === 'public') return images.value.filter((i) => i.isPublic)
  if (filter.value === 'private') return images.value.filter((i) => !i.isPublic)
  return images.value
})

function handleStart(imageId: string) {
  preselectedImageId.value = imageId
  showNewSandbox.value = true
}

async function handleCreateSandbox(
  payload: { imageId: string; ttlMinutes: number },
  done: (success: boolean) => void,
) {
  try {
    if (authStore.isAuthenticated) {
      await sandboxesStore.createSandbox(payload)
    } else {
      await sandboxesStore.createPublicDemo(payload)
    }
    toast.success('Sandbox wird gestartet')
    done(true)
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Fehler beim Starten der Sandbox'))
    done(false)
  }
}
</script>

<template>
  <div>
    <PageHeader title="Entdecken" subtitle="Starte eine Sandbox aus einer verfügbaren Vorlage." />

    <div class="space-y-6">
      <FilterBar v-model="filter" />

      <CardGridSkeleton v-if="loading" :count="6" />
      <PresetGrid
        v-else
        :images="filteredImages"
        @start="handleStart"
      />
    </div>

    <NewSandboxDialog
      v-model:open="showNewSandbox"
      :images="images"
      :preselected-image-id="preselectedImageId"
      @submit="handleCreateSandbox"
    />
  </div>
</template>
