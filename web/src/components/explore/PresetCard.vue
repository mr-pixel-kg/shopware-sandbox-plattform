<script setup lang="ts">
import type { Image } from '@/types'
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Play } from 'lucide-vue-next'

defineProps<{
  image: Image
}>()

const emit = defineEmits<{
  start: [imageId: string]
}>()
</script>

<template>
  <Card class="flex flex-col h-[200px]">
    <CardHeader class="pb-3">
      <div class="flex items-start justify-between">
        <CardTitle class="text-sm font-medium">{{ image.title || image.name }}</CardTitle>
      </div>
      <Badge variant="secondary" class="w-fit text-xs">
        {{ image.name }}:{{ image.tag }}
      </Badge>
    </CardHeader>
    <CardContent class="flex-1 pb-3">
      <p v-if="image.description" class="text-sm text-muted-foreground line-clamp-2">
        {{ image.description }}
      </p>
    </CardContent>
    <CardFooter>
      <Button size="sm" class="w-full" @click="emit('start', image.id)">
        <Play class="h-4 w-4 mr-1" />
        Starten
      </Button>
    </CardFooter>
  </Card>
</template>
