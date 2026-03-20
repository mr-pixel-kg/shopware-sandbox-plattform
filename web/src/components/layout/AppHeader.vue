<script setup lang="ts">
import { computed } from "vue";
import { RouterLink, useRouter } from "vue-router";
import { Boxes, ShieldCheck } from "lucide-vue-next";
import { useAuthStore } from "@/stores/auth";
import { Button } from "@/components/ui/button";

const auth = useAuthStore();
const router = useRouter();

const userLabel = computed(() => auth.user?.email ?? "Guest");

function logout() {
  auth.logout();
  router.push("/");
}
</script>

<template>
  <header class="border-b bg-background">
    <div class="mx-auto flex w-full max-w-7xl flex-col gap-4 px-4 py-4 md:flex-row md:items-center md:justify-between md:px-8">
      <div class="flex items-center gap-4">
        <div class="flex h-10 w-10 items-center justify-center rounded-md bg-primary text-primary-foreground">
          <Boxes class="h-5 w-5" />
        </div>
        <div>
          <RouterLink to="/" class="text-lg font-semibold tracking-tight">Shopshredder Sandbox Platform</RouterLink>
          <p class="text-sm text-muted-foreground">Public demos and internal sandbox management.</p>
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-3">
        <RouterLink to="/">
          <Button variant="ghost">Storefront</Button>
        </RouterLink>
        <template v-if="auth.isAuthenticated">
          <RouterLink to="/admin">
            <Button variant="secondary">
              <ShieldCheck class="mr-2 h-4 w-4" />
              Admin
            </Button>
          </RouterLink>
          <span class="rounded-md bg-secondary px-3 py-2 text-sm text-secondary-foreground">
            {{ userLabel }}
          </span>
          <Button variant="outline" @click="logout">Logout</Button>
        </template>
        <template v-else>
          <RouterLink to="/login">
            <Button>Login</Button>
          </RouterLink>
        </template>
      </div>
    </div>
  </header>
</template>
