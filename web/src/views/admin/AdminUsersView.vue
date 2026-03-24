<script setup lang="ts">
import { Trash2, UserPlus } from 'lucide-vue-next'
import { ref } from 'vue'
import { toast } from 'vue-sonner'

import ConfirmDialog from '@/components/modals/ConfirmDialog.vue'
import InviteUserDialog from '@/components/modals/InviteUserDialog.vue'
import PageHeader from '@/components/shared/PageHeader.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableEmpty,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { useWhitelist } from '@/composables/useWhitelist'
import { getApiErrorMessage } from '@/utils/error'
import { formatDateTime } from '@/utils/formatters'

const { pendingUsers, loading, add, remove } = useWhitelist()

const showInvite = ref(false)
const showConfirmDelete = ref(false)
const selectedUserId = ref<string | null>(null)

function requestDelete(id: string) {
  selectedUserId.value = id
  showConfirmDelete.value = true
}

async function handleInvite(
  payload: { email: string; role: 'admin' | 'user' },
  done: (success: boolean) => void,
) {
  try {
    await add(payload)
    toast.success('Benutzer wurde eingeladen')
    done(true)
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Fehler beim Einladen'))
    done(false)
  }
}

async function handleDelete() {
  if (!selectedUserId.value) return
  try {
    await remove(selectedUserId.value)
    toast.success('Einladung wurde entfernt')
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Fehler beim Entfernen'))
  }
}
</script>

<template>
  <div>
    <PageHeader title="Benutzer" subtitle="Benutzer und Zugriffsrechte verwalten.">
      <template #actions>
        <Button @click="showInvite = true">
          <UserPlus class="mr-1 h-4 w-4" />
          Benutzer einladen
        </Button>
      </template>
    </PageHeader>

    <Table class="table-fixed">
      <TableHeader>
        <TableRow>
          <TableHead class="w-[40%]">E-Mail</TableHead>
          <TableHead class="w-[20%]">Rolle</TableHead>
          <TableHead class="w-[25%]">Eingeladen am</TableHead>
          <TableHead class="w-[15%] text-right">Aktionen</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <template v-if="loading">
          <TableRow v-for="i in 3" :key="i" class="h-13">
            <TableCell><Skeleton class="h-4 w-40" /></TableCell>
            <TableCell><Skeleton class="h-5 w-16 rounded-full" /></TableCell>
            <TableCell><Skeleton class="h-4 w-32" /></TableCell>
            <TableCell><Skeleton class="ml-auto h-4 w-8" /></TableCell>
          </TableRow>
        </template>
        <TableEmpty v-else-if="pendingUsers.length === 0" :colspan="4">
          Keine ausstehenden Einladungen vorhanden.
        </TableEmpty>
        <TableRow v-for="user in pendingUsers" v-else :key="user.id">
          <TableCell class="font-medium">{{ user.email }}</TableCell>
          <TableCell>
            <Badge :variant="user.role === 'admin' ? 'default' : 'secondary'">
              {{ user.role === 'admin' ? 'Admin' : 'User' }}
            </Badge>
          </TableCell>
          <TableCell class="text-muted-foreground">
            {{ formatDateTime(user.createdAt) }}
          </TableCell>
          <TableCell class="text-right">
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger as-child>
                  <Button variant="ghost" size="icon-sm" @click="requestDelete(user.id)">
                    <Trash2 class="h-4 w-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Einladung entfernen</TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <InviteUserDialog v-model:open="showInvite" @invite="handleInvite" />

    <ConfirmDialog
      v-model:open="showConfirmDelete"
      title="Einladung entfernen"
      description="Möchtest du diese ausstehende Einladung wirklich entfernen? Der Benutzer kann sich dann nicht mehr registrieren."
      confirm-label="Entfernen"
      @confirm="handleDelete"
    />
  </div>
</template>
