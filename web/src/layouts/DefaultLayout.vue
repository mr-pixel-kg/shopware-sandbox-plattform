<script setup lang="ts">
import AppNavbar from '@/components/shared/AppNavbar.vue'
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <AppNavbar />
    <main class="mx-auto w-full max-w-6xl flex-1 px-6 py-8">
      <slot />
    </main>
  </div>
</template>
