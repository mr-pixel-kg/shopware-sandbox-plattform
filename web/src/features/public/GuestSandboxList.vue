<script setup lang="ts">
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { formatDateTime, relativeRemaining } from "@/lib/utils";
import type { SandboxRecord } from "@/types/api";

defineProps<{
  sandboxes: SandboxRecord[];
  deletingId?: string | null;
}>();

defineEmits<{
  remove: [id: string];
}>();

function progressValue(expiresAt?: string | null) {
  if (!expiresAt) return 0;

  const expiry = new Date(expiresAt).getTime();
  const created = expiry - 60 * 60 * 1000;
  const lifetime = expiry - created;
  const elapsed = Date.now() - created;

  return Math.max(0, Math.min(100, 100 - (elapsed / lifetime) * 100));
}
</script>

<template>
  <Card>
    <CardHeader class="space-y-4">
      <div class="flex items-center justify-between gap-4">
        <div>
          <CardTitle class="text-xl">Active demos</CardTitle>
          <p class="text-sm text-muted-foreground">Tracked automatically via guest session cookie.</p>
        </div>
        <Badge variant="secondary">{{ sandboxes.length }} active</Badge>
      </div>
    </CardHeader>

    <CardContent class="space-y-5">
      <div v-if="sandboxes.length === 0" class="rounded-md border border-dashed border-border bg-secondary/40 px-4 py-8 text-sm text-muted-foreground">
        You do not have any guest demo sandboxes yet.
      </div>

      <div v-else class="grid gap-4">
        <article v-for="sandbox in sandboxes" :key="sandbox.id" class="rounded-lg border bg-background p-4">
          <div class="space-y-3">
            <div class="flex items-start justify-between gap-3">
              <div class="space-y-2">
                <div class="flex flex-wrap items-center gap-2">
                  <h3 class="font-semibold">{{ sandbox.containerName }}</h3>
                  <Badge :variant="sandbox.status === 'running' ? 'default' : 'outline'">{{ sandbox.status }}</Badge>
                </div>
                <a :href="sandbox.url" target="_blank" rel="noreferrer" class="text-sm font-medium text-primary hover:underline">
                  {{ sandbox.url }}
                </a>
              </div>

              <Button variant="outline" size="sm" :disabled="deletingId === sandbox.id" @click="$emit('remove', sandbox.id)">
                Delete
              </Button>
            </div>

            <div class="space-y-2">
              <Progress :model-value="progressValue(sandbox.expiresAt)" />
              <div class="flex items-center justify-between text-xs text-muted-foreground">
                <span>{{ relativeRemaining(sandbox.expiresAt) }}</span>
                <span>{{ formatDateTime(sandbox.expiresAt) }}</span>
              </div>
            </div>
          </div>
        </article>
      </div>
    </CardContent>
  </Card>
</template>
