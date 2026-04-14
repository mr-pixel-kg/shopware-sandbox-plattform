<script setup lang="ts">
import { Play, Trash2 } from 'lucide-vue-next'
import { computed, ref } from 'vue'
import { toast } from 'vue-sonner'

import PresetGrid from '@/components/explore/PresetGrid.vue'
import SandboxCard from '@/components/explore/SandboxCard.vue'
import CardGridSkeleton from '@/components/shared/CardGridSkeleton.vue'
import PageHeader from '@/components/shared/PageHeader.vue'
import ShredderAnimation from '@/components/shared/ShredderAnimation.vue'
import { useImages } from '@/composables/useImages'
import { useSandboxes } from '@/composables/useSandboxes'
import { getApiErrorMessage } from '@/utils/error'
import { resolveAssetUrl } from '@/utils/formatters'
import { extractFieldValues } from '@/utils/metadata'

import type { CardAction } from '@/components/explore/ActionButton.vue'
import type { Image, Sandbox } from '@/types'

const { images, loading: imagesLoading } = useImages()
const {
  activeSandboxes,
  loading: sandboxesLoading,
  busyIds,
  createSandbox,
  deleteSandbox,
  removeSandbox,
  refresh: refreshSandboxes,
} = useSandboxes()
const shredderRefs = ref<Record<string, InstanceType<typeof ShredderAnimation>>>({})

const hasActiveSandboxes = computed(() => activeSandboxes.value.length > 0)

function getImageForSandbox(sandbox: Sandbox): Image | undefined {
  return images.value.find((i) => i.id === sandbox.imageId)
}

function getImageTitle(sandbox: Sandbox): string {
  if (sandbox.displayName) return sandbox.displayName
  const image = getImageForSandbox(sandbox)
  return image?.title || image?.name || sandbox.containerName
}

function getImageThumbnail(sandbox: Sandbox): string | undefined {
  return resolveAssetUrl(getImageForSandbox(sandbox)?.thumbnailUrl)
}

function sandboxExtraActions(sandbox: Sandbox): CardAction[] {
  if (sandbox.status !== 'running' && sandbox.status !== 'starting') return []
  return [
    {
      label: 'Stoppen',
      variant: 'destructive',
      icon: Trash2,
      size: 'icon',
      tooltip: 'Stoppen',
      loading: busyIds.value.has(sandbox.id),
      disabled: busyIds.value.has(sandbox.id),
      onClick: () => handleStopSandbox(sandbox),
    },
  ]
}

function presetExtraActions(image: Image): CardAction[] {
  return [
    {
      label: 'Demo starten',
      variant: 'default',
      icon: Play,
      loading: busyIds.value.has(image.id),
      disabled: busyIds.value.has(image.id),
      onClick: () => handleDemo(image.id),
    },
  ]
}

async function handleStopSandbox(sandbox: Sandbox) {
  if (busyIds.value.has(sandbox.id)) return
  busyIds.value.add(sandbox.id)
  try {
    const animationPromise = shredderRefs.value[sandbox.id]?.shred() ?? Promise.resolve()
    const apiPromise = deleteSandbox(sandbox.id, { skipRemove: true })
    await Promise.all([animationPromise, apiPromise])
    removeSandbox(sandbox.id)
    toast.success('Sandbox wird gestoppt')
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Fehler beim Stoppen'))
    refreshSandboxes()
  } finally {
    busyIds.value.delete(sandbox.id)
  }
}

function defaultsForImage(imageId: string): Record<string, string> | undefined {
  const image = images.value.find((i) => i.id === imageId)
  const defaults = extractFieldValues(image?.metadata)
  return Object.keys(defaults).length > 0 ? defaults : undefined
}

async function handleDemo(imageId: string) {
  if (busyIds.value.has(imageId)) return
  busyIds.value.add(imageId)
  try {
    const metadata = defaultsForImage(imageId)
    await createSandbox({ imageId, metadata })
    toast.success('Demo wird gestartet')
    refreshSandboxes()
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Fehler beim Starten der Demo'))
  } finally {
    busyIds.value.delete(imageId)
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
        <h3 class="text-muted-foreground mb-3 text-sm font-medium">Meine Sandboxes</h3>
        <CardGridSkeleton v-if="sandboxesLoading" :count="2" variant="sandbox" />
        <div
          v-else-if="hasActiveSandboxes"
          class="grid auto-rows-fr grid-cols-[repeat(auto-fill,minmax(320px,1fr))] gap-4"
        >
          <ShredderAnimation
            v-for="sandbox in activeSandboxes"
            :key="sandbox.id"
            :ref="
              (el: any) => {
                if (el) shredderRefs[sandbox.id] = el
              }
            "
          >
            <SandboxCard
              :sandbox="sandbox"
              :title="getImageTitle(sandbox)"
              :thumbnail-url="getImageThumbnail(sandbox)"
              :extra-actions="sandboxExtraActions(sandbox)"
              :state-reason="sandbox.stateReason"
              context="sandbox.card"
            />
          </ShredderAnimation>
        </div>
      </section>

      <section>
        <h3 class="text-muted-foreground mb-3 text-sm font-medium">Vorlagen</h3>
        <CardGridSkeleton v-if="imagesLoading" :count="6" />
        <PresetGrid v-else :images="images" :get-extra-actions="presetExtraActions" />
      </section>
    </div>
  </div>
</template>
