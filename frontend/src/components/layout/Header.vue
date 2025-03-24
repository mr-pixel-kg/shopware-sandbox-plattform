<script setup>
import logo from "@/assets/logo.png";
import { useAuthStore } from "@/stores/authStore";

const authStore = useAuthStore();

function handleLogout() {
  authStore.logout();
}

</script>

<template>
  <header>
    <nav class="bg-zinc-800 border-gray-200 px-4 lg:px-6 py-6">
      <div class="flex flex-wrap justify-between items-center mx-auto max-w-full">
        <RouterLink to="/" class="flex items-center">
          <img :src="logo" class="mr-3 h-6 sm:h-12" alt="mr. pixel Logo" />
          <span class="self-center text-xl font-semibold whitespace-nowrap text-white">Sandbox Environment</span>
        </RouterLink>
        <div class="flex items-center lg:order-2">
          <RouterLink to="/login" class="text-white bg-zinc-800 hover:bg-zinc-700 font-medium rounded-lg text-base px-4 lg:px-5 py-2 lg:py-2.5 mr-2" v-if="!authStore.isAuthenticated">Login</RouterLink>
          <RouterLink to="/admin" class="text-white bg-zinc-800 hover:bg-zinc-700 font-medium rounded-lg text-base px-4 lg:px-5 py-2 lg:py-2.5 mr-2" v-if="authStore.isAuthenticated">Admin</RouterLink>
          <a class="text-white bg-zinc-800 hover:bg-zinc-700 font-medium rounded-lg text-base px-4 lg:px-5 py-2 lg:py-2.5 mr-2 hover:cursor-pointer" v-if="authStore.isAuthenticated" @click="handleLogout">Logout</a>
        </div>
      </div>
    </nav>
  </header>
</template>

<style scoped>

</style>