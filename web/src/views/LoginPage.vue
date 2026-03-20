<script setup lang="ts">
import { reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import AppHeader from "@/components/layout/AppHeader.vue";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
const route = useRoute();
const router = useRouter();
const form = reactive({
  email: "",
  password: "",
});
const pending = ref(false);
const error = ref<string | null>(null);

async function submit() {
  pending.value = true;
  error.value = null;

  try {
    await auth.login(form.email, form.password);
    const redirect = typeof route.query.redirect === "string" ? route.query.redirect : "/admin";
    await router.push(redirect);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Login failed";
  } finally {
    pending.value = false;
  }
}
</script>

<template>
  <div class="app-shell">
    <AppHeader />

    <div class="mx-auto flex w-full max-w-5xl flex-1 items-center justify-center">
      <div class="grid w-full gap-6 lg:grid-cols-[minmax(0,1fr)_420px]">
        <section class="hidden rounded-xl border bg-card px-8 py-10 lg:block">
          <span class="inline-flex rounded-md bg-primary/15 px-3 py-1 text-sm font-semibold text-primary">Employee access</span>
          <h1 class="mt-6 text-4xl font-extrabold tracking-tight">Manage images, internal sandboxes and audit logs.</h1>
          <p class="mt-4 text-base text-muted-foreground">
            Sign in with your employee account to publish image templates, create editable internal sandboxes and inspect the full platform activity.
          </p>
        </section>

        <Card class="self-center">
          <CardHeader>
            <CardTitle>Login</CardTitle>
            <CardDescription>Use your employee credentials to access the admin workspace.</CardDescription>
          </CardHeader>
          <CardContent class="space-y-5">
            <div v-if="error" class="rounded-md border border-danger/20 bg-danger/10 px-4 py-3 text-sm text-danger">
              {{ error }}
            </div>

            <form class="space-y-4" @submit.prevent="submit">
              <div class="space-y-2">
                <Label for="email">Email</Label>
                <Input id="email" v-model="form.email" type="email" placeholder="dev@shopshredder.de" />
              </div>
              <div class="space-y-2">
                <Label for="password">Password</Label>
                <Input id="password" v-model="form.password" type="password" placeholder="••••••••" />
              </div>
              <Button class="w-full" type="submit" :disabled="pending">Login</Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  </div>
</template>
