<script setup lang="ts">
import { ref, watch } from 'vue'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Loader2 } from 'lucide-vue-next'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  submit: [
    payload: { name: string; tag: string; title: string; description: string; isPublic: boolean },
    done: (success: boolean) => void,
  ]
}>()

const name = ref('')
const tag = ref('')
const title = ref('')
const description = ref('')
const isPublic = ref(true)
const busy = ref(false)

function resetState() {
  name.value = ''
  tag.value = ''
  title.value = ''
  description.value = ''
  isPublic.value = true
  busy.value = false
}

watch(
  () => props.open,
  (open) => {
    if (open) resetState()
  },
)

function handleSubmit() {
  if (!name.value || !tag.value) return
  busy.value = true
  emit(
    'submit',
    {
      name: name.value,
      tag: tag.value,
      title: title.value,
      description: description.value,
      isPublic: isPublic.value,
    },
    (success: boolean) => {
      busy.value = false
      if (success) {
        emit('update:open', false)
      }
    },
  )
}
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="sm:max-w-[500px]">
      <DialogHeader>
        <DialogTitle>Vorlage hinzufügen</DialogTitle>
        <DialogDescription>Füge ein neues Docker-Image als Sandbox-Vorlage hinzu.</DialogDescription>
      </DialogHeader>
      <form @submit.prevent="handleSubmit" class="grid gap-4 py-4">
        <div class="grid gap-2">
          <Label for="image-name">Image Name</Label>
          <Input id="image-name" v-model="name" placeholder="dockware/dev" required :disabled="busy" />
        </div>
        <div class="grid gap-2">
          <Label for="image-tag">Tag</Label>
          <Input id="image-tag" v-model="tag" placeholder="latest" required :disabled="busy" />
        </div>
        <div class="grid gap-2">
          <Label for="image-title">Titel</Label>
          <Input id="image-title" v-model="title" placeholder="Leere Installation" :disabled="busy" />
        </div>
        <div class="grid gap-2">
          <Label for="image-description">Beschreibung</Label>
          <Textarea id="image-description" v-model="description" placeholder="Beschreibung der Vorlage..." :disabled="busy" />
        </div>
        <div class="flex items-center justify-between">
          <Label for="image-public">Öffentlich sichtbar</Label>
          <Switch id="image-public" v-model="isPublic" :disabled="busy" />
        </div>
        <DialogFooter class="pt-2">
          <Button type="button" variant="outline" :disabled="busy" @click="emit('update:open', false)">Abbrechen</Button>
          <Button type="submit" :disabled="!name || !tag || busy">
            <Loader2 v-if="busy" class="h-4 w-4 animate-spin mr-1" />
            {{ busy ? 'Wird hinzugefügt...' : 'Hinzufügen' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
