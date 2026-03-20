<script setup lang="ts">
import { ref, computed } from 'vue'
import { useImages } from '@/composables/useImages'
import { useSandboxes } from '@/composables/useSandboxes'
import { useAuthStore } from '@/stores/auth.store'
import { getApiErrorMessage } from '@/utils/error'
import { toast } from 'vue-sonner'
import { ExternalLink, Play, Square } from 'lucide-vue-next'
import type { Image, Sandbox } from '@/types'
import type { CardAction } from '@/components/explore/ActionButton.vue'
import PageHeader from '@/components/shared/PageHeader.vue'
import PresetGrid from '@/components/explore/PresetGrid.vue'
import CardGridSkeleton from '@/components/shared/CardGridSkeleton.vue'
import SandboxCard from '@/components/explore/SandboxCard.vue'
import type { MetadataGroup } from '@/components/explore/SandboxCard.vue'

const { images, loading: imagesLoading } = useImages()
const {
  activeSandboxes,
  loading: sandboxesLoading,
  createPublicDemo,
  createSandbox,
  deleteSandbox,
  refresh: refreshSandboxes,
} = useSandboxes()
const authStore = useAuthStore()

const creatingImageId = ref<string>()
const stoppingSandboxId = ref<string>()

const hasActiveSandboxes = computed(() => activeSandboxes.value.length > 0)

function getImageTitle(sandbox: Sandbox): string {
  const image = images.value.find((i) => i.id === sandbox.imageId)
  return image?.title || image?.name || sandbox.containerName
}

// TODO: Replace with dynamic schema from API
const sandboxActionsMap = computed(() => {
  const map: Record<string, CardAction[]> = {}
  for (const sandbox of activeSandboxes.value) {
    const actions: CardAction[] = []
    if (sandbox.url && sandbox.status === 'running') {
      actions.push({
        label: 'Öffnen',
        href: sandbox.url,
        variant: 'default',
        icon: ExternalLink,
      })
    }
    if (sandbox.status === 'running' || sandbox.status === 'starting') {
      actions.push({
        label: 'Stoppen',
        variant: 'destructive',
        icon: Square,
        loading: stoppingSandboxId.value === sandbox.id,
        disabled: !!stoppingSandboxId.value && stoppingSandboxId.value !== sandbox.id,
        onClick: () => handleStopSandbox(sandbox),
      })
    }
    map[sandbox.id] = actions
  }
  return map
})

// TODO: Replace with dynamic schema from API
const sandboxMetadataMap = computed(() => {
  const map: Record<string, MetadataGroup[]> = {}
  for (const sandbox of activeSandboxes.value) {
    map[sandbox.id] = [
      {
        title: 'Zugangsdaten',
        fields: [
          { label: 'Benutzername', value: 'admin' },
          { label: 'Passwort', value: 'shopware', secret: true },
        ],
      },
    ]
  }
  return map
})

// TODO: Replace with dynamic schema from API
function getPresetActions(image: Image): CardAction[] {
  return [
    {
      label: 'Demo starten',
      variant: 'default',
      icon: Play,
      loading: creatingImageId.value === image.id,
      disabled: !!creatingImageId.value && creatingImageId.value !== image.id,
      onClick: () => handleDemo(image.id),
    },
  ]
}

async function handleStopSandbox(sandbox: Sandbox) {
  if (stoppingSandboxId.value) return
  stoppingSandboxId.value = sandbox.id
  try {
    if (authStore.isAuthenticated) {
      await deleteSandbox(sandbox.id)
    } else {
      const { sandboxesApi } = await import('@/api')
      await sandboxesApi.removeGuest(sandbox.id)
      refreshSandboxes()
    }
    toast.success('Sandbox wird gestoppt')
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Fehler beim Stoppen'))
  } finally {
    stoppingSandboxId.value = undefined
  }
}

async function handleDemo(imageId: string) {
  if (creatingImageId.value) return
  creatingImageId.value = imageId
  try {
    if (authStore.isAuthenticated) {
      await createSandbox({ imageId })
    } else {
      await createPublicDemo({ imageId })
    }
    toast.success('Demo wird gestartet')
    refreshSandboxes()
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Fehler beim Starten der Demo'))
  } finally {
    creatingImageId.value = undefined
  }
}
</script>

<template>
  <div>
    <PageHeader
      title="Entdecken"
      subtitle="Starte eine Demo aus einer verfügbaren Vorlage — kein Account nötig."
    />

    <div class="space-y-8">
      <section v-if="hasActiveSandboxes || sandboxesLoading">
        <h3 class="text-sm font-medium text-muted-foreground mb-3">Meine Sandboxes</h3>
        <CardGridSkeleton v-if="sandboxesLoading" :count="2" />
        <div
          v-else-if="hasActiveSandboxes"
          class="grid grid-cols-[repeat(auto-fill,minmax(240px,1fr))] gap-4"
        >
          <SandboxCard
            v-for="sandbox in activeSandboxes"
            :key="sandbox.id"
            :sandbox="sandbox"
            :title="getImageTitle(sandbox)"
            :actions="sandboxActionsMap[sandbox.id]"
            :metadata="sandboxMetadataMap[sandbox.id]"
          />
        </div>
      </section>

      <section>
        <h3 class="text-sm font-medium text-muted-foreground mb-3">Vorlagen</h3>
        <CardGridSkeleton v-if="imagesLoading" :count="6" />
        <PresetGrid
          v-else
          :images="images"
          :get-actions="getPresetActions"
        />
      </section>
    </div>
  </div>
</template>
