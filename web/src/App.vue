<script setup lang="ts">
import 'vue-sonner/style.css'
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import { Toaster } from '@/components/ui/sonner'
import AuthLayout from '@/layouts/AuthLayout.vue'
import DefaultLayout from '@/layouts/DefaultLayout.vue'

const route = useRoute()
const layout = computed(() => (route.meta.layout === 'auth' ? AuthLayout : DefaultLayout))
</script>

<template>
  <component :is="layout">
    <RouterView />
  </component>
  <Toaster rich-colors />
</template>
