<script setup>
import logo from './assets/logo.png'
import Toast from 'primevue/toast';
</script>

<template>
  <header class="header">
    <div class="logo-wrapper">
      <img :src="logo" alt="mr. pixel Logo" class="logo" />
    </div>
    <nav>
      <RouterLink to="/">Go to Home</RouterLink>
      <RouterLink to="/login">Go to Login</RouterLink>
      <RouterLink to="/admin">Go to Admin</RouterLink>
    </nav>
    <Toast />
  </header>

  <main>
    <RouterView />
  </main>
</template>

<style scoped>
.header {
  background-color: #242424;
  padding: 20px;
}
.logo-wrapper {

}
.logo {
  width: 250px;
}
</style>
