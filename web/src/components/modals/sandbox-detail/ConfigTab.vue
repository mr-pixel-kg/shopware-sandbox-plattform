<script setup lang="ts">
import { computed } from 'vue'

import MetadataSection from '@/components/metadata/MetadataSection.vue'
import { extractFieldValues } from '@/utils/metadata'

import type { Sandbox } from '@/types'

const props = defineProps<{
  sandbox: Sandbox
}>()

const values = computed(() => extractFieldValues(props.sandbox.metadata))
const metadata = computed(() => props.sandbox.metadata ?? [])

defineExpose({ resetRevealed: () => {} })
</script>

<template>
  <MetadataSection
    :metadata="metadata"
    context="sandbox.details"
    :model-value="values"
    view="form"
    disabled
    :actions-disabled="sandbox.status !== 'running'"
  />
</template>
