<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Textarea } from "@/components/ui/textarea";
import { formatDateTime, relativeRemaining } from "@/lib/utils";
import type { CreateImagePayload, CreateSandboxPayload, ImageRecord, SandboxRecord } from "@/types/api";

const props = defineProps<{
  images: ImageRecord[];
  sandboxes: SandboxRecord[];
  creating: boolean;
  deletingId?: string | null;
  snapshottingId?: string | null;
}>();

const emit = defineEmits<{
  create: [payload: CreateSandboxPayload];
  remove: [id: string];
  snapshot: [sandboxId: string, payload: CreateImagePayload];
}>();

const createForm = reactive({
  imageId: "",
  ttlMinutes: 120,
});

const snapshotForm = reactive({
  name: "",
  tag: "",
  title: "",
  description: "",
  thumbnailUrl: "",
  isPublic: false,
});

const snapshotTarget = ref<string | null>(null);

const selectableImages = computed(() => props.images.map((image) => ({
  value: image.id,
  label: `${image.name}:${image.tag}`,
})));

function createSandbox() {
  emit("create", {
    imageId: createForm.imageId,
    ttlMinutes: createForm.ttlMinutes || null,
  });
}

function openSnapshot(id: string) {
  snapshotTarget.value = id;
}

function submitSnapshot() {
  if (!snapshotTarget.value) return;

  emit("snapshot", snapshotTarget.value, {
    ...snapshotForm,
    title: snapshotForm.title || null,
    description: snapshotForm.description || null,
    thumbnailUrl: snapshotForm.thumbnailUrl || null,
  });
}
</script>

<template>
  <div class="grid gap-6 xl:grid-cols-[360px_minmax(0,1fr)]">
    <Card>
      <CardHeader>
        <CardTitle>Start internal sandbox</CardTitle>
        <CardDescription>Create editable employee sandboxes from registered images.</CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <div class="space-y-2">
          <Label>Image</Label>
          <Select v-model="createForm.imageId">
            <SelectTrigger>
              <SelectValue placeholder="Select an image" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="option in selectableImages" :key="option.value" :value="option.value">
                {{ option.label }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div class="space-y-2">
          <Label for="ttl-minutes">TTL in minutes</Label>
          <Input id="ttl-minutes" v-model="createForm.ttlMinutes" type="number" placeholder="120" />
        </div>
        <Button class="w-full" :disabled="creating || !createForm.imageId" @click="createSandbox">Create sandbox</Button>
      </CardContent>
    </Card>

    <Card>
      <CardHeader class="flex flex-row items-center justify-between space-y-0">
        <div class="space-y-1">
          <CardTitle>Running sandboxes</CardTitle>
          <CardDescription>Includes guest demos and internal employee environments.</CardDescription>
        </div>
        <Badge variant="secondary">{{ sandboxes.length }} active</Badge>
      </CardHeader>
      <CardContent>
        <div class="overflow-hidden rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Container</TableHead>
                <TableHead>Owner</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>URL</TableHead>
                <TableHead>Expires</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="sandbox in sandboxes" :key="sandbox.id" class="align-top">
                <TableCell class="font-medium">{{ sandbox.containerName }}</TableCell>
                <TableCell>
                  <Badge :variant="sandbox.createdByUserId ? 'default' : 'outline'">{{ sandbox.createdByUserId ? "Employee" : "Guest" }}</Badge>
                </TableCell>
                <TableCell>
                  <Badge :variant="sandbox.status === 'running' ? 'default' : 'outline'">{{ sandbox.status }}</Badge>
                </TableCell>
                <TableCell>
                  <a :href="sandbox.url" target="_blank" rel="noreferrer" class="text-sm font-medium text-primary hover:underline">
                    {{ sandbox.url }}
                  </a>
                </TableCell>
                <TableCell class="text-muted-foreground">
                  <div>{{ formatDateTime(sandbox.expiresAt) }}</div>
                  <div class="text-xs">{{ relativeRemaining(sandbox.expiresAt) }}</div>
                </TableCell>
                <TableCell>
                  <div class="flex flex-wrap gap-2">
                    <Button variant="secondary" size="sm" @click="openSnapshot(sandbox.id)">Snapshot</Button>
                    <Button variant="outline" size="sm" :disabled="deletingId === sandbox.id" @click="$emit('remove', sandbox.id)">
                      Delete
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  </div>

  <Card v-if="snapshotTarget" class="mt-6">
    <CardHeader class="flex flex-row items-center justify-between space-y-0">
      <div class="space-y-1">
        <CardTitle>Create snapshot image</CardTitle>
        <CardDescription>Commit the selected sandbox into a reusable image template.</CardDescription>
      </div>
      <Button variant="ghost" @click="snapshotTarget = null">Close</Button>
    </CardHeader>

    <CardContent class="space-y-4">
      <div class="grid gap-4 md:grid-cols-2">
        <div class="space-y-2">
          <Label for="snapshot-name">Name</Label>
          <Input id="snapshot-name" v-model="snapshotForm.name" placeholder="ghcr.io/shopshredder/shopware-custom" />
        </div>
        <div class="space-y-2">
          <Label for="snapshot-tag">Tag</Label>
          <Input id="snapshot-tag" v-model="snapshotForm.tag" placeholder="demo-v2" />
        </div>
        <div class="space-y-2 md:col-span-2">
          <Label for="snapshot-title">Title</Label>
          <Input id="snapshot-title" v-model="snapshotForm.title" placeholder="Shopware Custom Demo V2" />
        </div>
        <div class="space-y-2 md:col-span-2">
          <Label for="snapshot-description">Description</Label>
          <Textarea id="snapshot-description" v-model="snapshotForm.description" placeholder="Snapshot taken from a configured employee sandbox." />
        </div>
        <div class="space-y-2 md:col-span-2">
          <Label for="snapshot-thumbnail">Thumbnail URL</Label>
          <Input id="snapshot-thumbnail" v-model="snapshotForm.thumbnailUrl" placeholder="https://..." />
        </div>
        <label class="flex items-center gap-3 rounded-md border bg-muted/30 px-3 py-3 text-sm md:col-span-2">
          <Checkbox :checked="snapshotForm.isPublic" @update:checked="snapshotForm.isPublic = !!$event" />
          Publish snapshot to the public storefront
        </label>
      </div>

      <Button :disabled="snapshottingId === snapshotTarget" @click="submitSnapshot">Commit snapshot</Button>
    </CardContent>
  </Card>
</template>
