<script setup lang="ts">
import { reactive } from "vue";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Textarea } from "@/components/ui/textarea";
import type { CreateImagePayload, ImageRecord } from "@/types/api";

defineProps<{
  images: ImageRecord[];
  creating: boolean;
  deletingId?: string | null;
}>();

const emit = defineEmits<{
  create: [payload: CreateImagePayload];
  remove: [id: string];
}>();

const form = reactive({
  name: "",
  tag: "latest",
  title: "",
  description: "",
  thumbnailUrl: "",
  isPublic: true,
});

function submit() {
  emit("create", {
    ...form,
    title: form.title || null,
    description: form.description || null,
    thumbnailUrl: form.thumbnailUrl || null,
  });
}
</script>

<template>
  <div class="grid gap-6 xl:grid-cols-[360px_minmax(0,1fr)]">
    <Card>
      <CardHeader>
        <CardTitle>Register image</CardTitle>
        <CardDescription>The backend pulls the Docker image if it is not already available locally.</CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <div class="space-y-2">
          <Label for="image-name">Name</Label>
          <Input id="image-name" v-model="form.name" placeholder="ghcr.io/shopshredder/shopware-demo" />
        </div>
        <div class="space-y-2">
          <Label for="image-tag">Tag</Label>
          <Input id="image-tag" v-model="form.tag" placeholder="latest" />
        </div>
        <div class="space-y-2">
          <Label for="image-title">Title</Label>
          <Input id="image-title" v-model="form.title" placeholder="Shopware Demo" />
        </div>
        <div class="space-y-2">
          <Label for="image-description">Description</Label>
          <Textarea id="image-description" v-model="form.description" placeholder="Public demo image for storefront previews" />
        </div>
        <div class="space-y-2">
          <Label for="image-thumbnail">Thumbnail URL</Label>
          <Input id="image-thumbnail" v-model="form.thumbnailUrl" placeholder="https://..." />
        </div>
        <label class="flex items-center gap-3 rounded-md border bg-muted/30 px-3 py-3 text-sm">
          <Checkbox :checked="form.isPublic" @update:checked="form.isPublic = !!$event" />
          Visible on the public storefront
        </label>
        <Button class="w-full" :disabled="creating" @click="submit">Create image</Button>
      </CardContent>
    </Card>

    <Card>
      <CardHeader class="flex flex-row items-center justify-between space-y-0">
        <div class="space-y-1">
          <CardTitle>Image catalog</CardTitle>
          <CardDescription>All registered templates managed by the internal team.</CardDescription>
        </div>
        <Badge variant="secondary">{{ images.length }} images</Badge>
      </CardHeader>
      <CardContent>
        <div class="overflow-hidden rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Image</TableHead>
                <TableHead>Title</TableHead>
                <TableHead>Public</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="image in images" :key="image.id">
                <TableCell class="font-medium">{{ image.name }}:{{ image.tag }}</TableCell>
                <TableCell>{{ image.title || "Untitled" }}</TableCell>
                <TableCell>
                  <Badge :variant="image.isPublic ? 'default' : 'outline'">{{ image.isPublic ? "Yes" : "No" }}</Badge>
                </TableCell>
                <TableCell>
                  <Button variant="outline" size="sm" :disabled="deletingId === image.id" @click="$emit('remove', image.id)">
                    Delete
                  </Button>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
