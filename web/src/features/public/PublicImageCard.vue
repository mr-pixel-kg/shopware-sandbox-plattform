<script setup lang="ts">
import { computed } from "vue";
import { ArrowUpRight, Clock3 } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import type { ImageRecord } from "@/types/api";

const props = defineProps<{
  image: ImageRecord;
  busy?: boolean;
}>();

const emit = defineEmits<{
  demo: [imageId: string];
}>();

const title = computed(() => props.image.title || `${props.image.name}:${props.image.tag}`);
</script>

<template>
  <Card class="flex h-full flex-col overflow-hidden">
    <div class="overflow-hidden bg-secondary">
      <img
        v-if="image.thumbnailUrl"
        :src="image.thumbnailUrl"
        :alt="title"
        class="h-40 w-full object-cover"
      />
      <div v-else class="flex h-40 items-center justify-center bg-secondary text-center text-sm text-muted-foreground">
        No thumbnail configured
      </div>
    </div>

    <CardHeader class="space-y-2">
      <CardTitle class="text-base">{{ title }}</CardTitle>
      <p class="text-sm text-muted-foreground">
        {{ image.description || "No description available yet for this image." }}
      </p>
    </CardHeader>

    <CardContent class="mt-auto">
      <div class="flex items-center justify-between text-xs text-muted-foreground">
        <span class="inline-flex items-center gap-1">
          <Clock3 class="h-3.5 w-3.5" />
          Demo lifetime managed by the backend
        </span>
        <span>{{ image.tag }}</span>
      </div>
    </CardContent>

    <CardFooter>
      <Button class="w-full" :disabled="busy" @click="emit('demo', image.id)">
        <ArrowUpRight class="mr-2 h-4 w-4" />
        Start demo
      </Button>
    </CardFooter>
  </Card>
</template>
