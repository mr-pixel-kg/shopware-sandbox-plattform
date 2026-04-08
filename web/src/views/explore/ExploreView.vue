<script setup lang="ts">
import { ExternalLink, Play, Trash2 } from 'lucide-vue-next'
import { computed, ref } from 'vue'
import { toast } from 'vue-sonner'

import PresetGrid from '@/components/explore/PresetGrid.vue'
import SandboxCard from '@/components/explore/SandboxCard.vue'
import CardGridSkeleton from '@/components/shared/CardGridSkeleton.vue'
import PageHeader from '@/components/shared/PageHeader.vue'
import ShredderAnimation from '@/components/shared/ShredderAnimation.vue'
import { useImages } from '@/composables/useImages'
import { useSandboxes } from '@/composables/useSandboxes'
import { useAuthStore } from '@/stores/auth.store'
import { getApiErrorMessage } from '@/utils/error'
import { resolveAssetUrl } from '@/utils/formatters'
import { resolveIcon } from '@/utils/icons'

import type { CardAction } from '@/components/explore/ActionButton.vue'
import type { MetadataGroup } from '@/components/explore/SandboxCard.vue'
import type { Image, MetadataItem, Sandbox } from '@/types'

const { images, loading: imagesLoading } = useImages()
const {
  activeSandboxes,
  loading: sandboxesLoading,
  busyIds,
  createDemo,
  createSandbox,
  deleteSandbox,
  removeSandbox,
  refresh: refreshSandboxes,
} = useSandboxes()
const authStore = useAuthStore()
const shredderRefs = ref<Record<string, InstanceType<typeof ShredderAnimation>>>({})

const hasActiveSandboxes = computed(() => activeSandboxes.value.length > 0)

function isSandboxReachable(sandbox: Sandbox): boolean {
  return sandbox.status === 'running'
}

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

function getSandboxMetadata(sandbox: Sandbox): MetadataItem[] {
  if (!sandbox.metadata || !Array.isArray(sandbox.metadata)) return []
  return sandbox.metadata
}

function resolveActionUrl(template: string, sandbox: Sandbox): string {
  return template
    .replace(/\{\{\.URL}}/g, sandbox.url)
    .replace(/\{\{\.Hostname}}/g, sandbox.url.replace(/^https?:\/\//, ''))
    .replace(/\{\{\.SandboxID}}/g, sandbox.id)
}

function showOnSandbox(item: MetadataItem): boolean {
  return !item.show || item.show === 'sandbox' || item.show === 'both'
}

function showOnTemplate(item: MetadataItem): boolean {
  return item.show === 'template' || item.show === 'both'
}

function isDisabledByCondition(item: MetadataItem, sandbox: Sandbox): boolean {
  if (item.condition === 'ready') return !isSandboxReachable(sandbox)
  return false
}

function metadataToAction(item: MetadataItem, sandbox: Sandbox): CardAction {
  return {
    label: item.label,
    href: resolveActionUrl(item.value ?? '', sandbox),
    variant: (item.variant as CardAction['variant']) ?? 'outline',
    icon: resolveIcon(item.icon) ?? ExternalLink,
    disabled: isDisabledByCondition(item, sandbox),
    size: (item.size as CardAction['size']) ?? 'default',
    tooltip: item.size === 'icon' ? item.label : undefined,
  }
}

const sandboxActionsMap = computed(() => {
  const map: Record<string, CardAction[]> = {}
  for (const sandbox of activeSandboxes.value) {
    const meta = getSandboxMetadata(sandbox)
    const actionItems = meta.filter((m) => m.type === 'action' && showOnSandbox(m))
    const actions: CardAction[] = []

    if (actionItems.length > 0) {
      for (const item of actionItems) {
        actions.push(metadataToAction(item, sandbox))
      }
    } else {
      actions.push({
        label: 'Öffnen',
        href: sandbox.url,
        variant: 'default',
        icon: ExternalLink,
        disabled: !isSandboxReachable(sandbox),
      })
    }

    if (sandbox.status === 'running' || sandbox.status === 'starting') {
      actions.push({
        label: 'Stoppen',
        variant: 'destructive',
        icon: Trash2,
        size: 'icon',
        tooltip: 'Stoppen',
        loading: busyIds.value.has(sandbox.id),
        disabled: busyIds.value.has(sandbox.id),
        onClick: () => handleStopSandbox(sandbox),
      })
    }
    map[sandbox.id] = actions
  }
  return map
})

function metadataToField(item: MetadataItem, sandbox?: Sandbox) {
  return {
    label: item.label,
    value: item.value || '',
    secret: item.input === 'password',
    icon: item.icon,
    loading: sandbox ? isDisabledByCondition(item, sandbox) : false,
  }
}

const sandboxMetadataMap = computed(() => {
  const map: Record<string, MetadataGroup[]> = {}
  for (const sandbox of activeSandboxes.value) {
    const meta = getSandboxMetadata(sandbox)
    const groups: MetadataGroup[] = []

    const configItems = meta.filter(
      (m) => (m.type === 'field' || m.type === 'setting') && showOnSandbox(m),
    )
    if (configItems.length > 0) {
      groups.push({
        title: 'Konfiguration',
        fields: configItems.map((m) => metadataToField(m, sandbox)),
      })
    }

    const infoItems = meta.filter((m) => m.type === 'info' && showOnSandbox(m))
    if (infoItems.length > 0) {
      groups.push({ title: 'Details', fields: infoItems.map((m) => metadataToField(m, sandbox)) })
    }

    if (groups.length > 0) map[sandbox.id] = groups
  }
  return map
})

function getPresetMetadata(image: Image): MetadataGroup[] {
  const meta = image.metadata ?? []
  const groups: MetadataGroup[] = []

  const configItems = meta.filter(
    (m) => (m.type === 'field' || m.type === 'setting') && showOnTemplate(m),
  )
  if (configItems.length > 0) {
    groups.push({ title: 'Konfiguration', fields: configItems.map((m) => metadataToField(m)) })
  }

  const infoItems = meta.filter((m) => m.type === 'info' && showOnTemplate(m))
  if (infoItems.length > 0) {
    groups.push({ title: 'Details', fields: infoItems.map((m) => metadataToField(m)) })
  }

  return groups
}

function getPresetActions(image: Image): CardAction[] {
  const actions: CardAction[] = []

  const templateActions = (image.metadata ?? []).filter(
    (m) => m.type === 'action' && showOnTemplate(m),
  )
  for (const item of templateActions) {
    actions.push({
      label: item.label,
      href: item.value,
      variant: (item.variant as CardAction['variant']) ?? 'outline',
      icon: resolveIcon(item.icon) ?? ExternalLink,
      size: (item.size as CardAction['size']) ?? 'default',
      tooltip: item.size === 'icon' ? item.label : undefined,
    })
  }

  actions.push({
    label: 'Demo starten',
    variant: 'default',
    icon: Play,
    loading: busyIds.value.has(image.id),
    disabled: busyIds.value.has(image.id),
    onClick: () => handleDemo(image.id),
  })

  return actions
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

function getMetadataDefaults(imageId: string): Record<string, string> | undefined {
  const image = images.value.find((i) => i.id === imageId)
  const meta = image?.metadata ?? []
  const defaults: Record<string, string> = {}
  for (const item of meta) {
    if ((item.type === 'field' || item.type === 'setting') && item.value) {
      defaults[item.key] = item.value
    }
  }
  return Object.keys(defaults).length > 0 ? defaults : undefined
}

async function handleDemo(imageId: string) {
  if (busyIds.value.has(imageId)) return
  busyIds.value.add(imageId)
  try {
    const metadata = getMetadataDefaults(imageId)
    if (authStore.isAuthenticated) {
      await createSandbox({ imageId, metadata })
    } else {
      await createDemo({ imageId })
    }
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
          class="grid auto-rows-[460px] grid-cols-[repeat(auto-fill,minmax(320px,1fr))] gap-4"
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
              :actions="sandboxActionsMap[sandbox.id]"
              :metadata="sandboxMetadataMap[sandbox.id]"
              :state-reason="sandbox.stateReason"
            />
          </ShredderAnimation>
        </div>
      </section>

      <section>
        <h3 class="text-muted-foreground mb-3 text-sm font-medium">Vorlagen</h3>
        <CardGridSkeleton v-if="imagesLoading" :count="6" />
        <PresetGrid
          v-else
          :images="images"
          :get-actions="getPresetActions"
          :get-metadata="getPresetMetadata"
        />
      </section>
    </div>
  </div>
</template>
