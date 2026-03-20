<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import AuditLogManager from "@/features/admin/AuditLogManager.vue";
import { api } from "@/lib/api";
import { useAuthStore } from "@/stores/auth";
import type { AuditLogRecord } from "@/types/api";

const auth = useAuthStore();

const logs = ref<AuditLogRecord[]>([]);
const loading = ref(true);
const error = ref<string | null>(null);

const token = computed(() => auth.token);

async function load() {
  if (!token.value) return;

  loading.value = true;
  error.value = null;

  try {
    logs.value = await api.getAuditLogs(token.value);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Could not load audit log";
  } finally {
    loading.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="space-y-6">
    <div class="space-y-1">
      <h1 class="text-2xl font-semibold tracking-tight">Audit Log</h1>
      <p class="text-sm text-muted-foreground">Review backend activity and operational events.</p>
    </div>

    <div v-if="error" class="rounded-md border border-destructive/20 bg-destructive/10 px-4 py-3 text-sm text-destructive">
      {{ error }}
    </div>

    <div v-if="loading" class="rounded-lg border bg-background px-4 py-8 text-sm text-muted-foreground">
      Loading audit log...
    </div>

    <AuditLogManager v-else :logs="logs" />
  </div>
</template>
