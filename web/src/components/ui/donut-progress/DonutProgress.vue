<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue?: number
  }>(),
  {
    modelValue: 0,
  },
)

const cx = 24
const cy = 24
const r = 20
const circumference = 2 * Math.PI * r
const dashoffset = computed(() => circumference * ((100 - Math.min(Math.max(props.modelValue, 0), 100)) / 100))
</script>

<template>
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="24"
    height="24"
    viewBox="0 0 48 48"
    fill="none"
    shape-rendering="geometricPrecision"
    style="transform: rotate(-90deg) translateZ(0); will-change: transform; backface-visibility: hidden;"
  >
    <!-- Track -->
    <circle
      :cx="cx"
      :cy="cy"
      :r="r"
      stroke="currentColor"
      stroke-width="4"
      opacity="0.2"
    />
    <!-- Progress arc -->
    <circle
      :cx="cx"
      :cy="cy"
      :r="r"
      stroke="currentColor"
      stroke-width="4"
      stroke-linecap="round"
      :stroke-dasharray="circumference"
      :stroke-dashoffset="dashoffset"
      class="transition-[stroke-dashoffset] duration-300 ease-out"
    />
  </svg>
</template>
