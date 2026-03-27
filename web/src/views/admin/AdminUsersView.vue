<script setup lang="ts">
import { Pencil, Trash2, UserPlus, Users } from 'lucide-vue-next'
import { computed, ref } from 'vue'
import { toast } from 'vue-sonner'

import ConfirmDialog from '@/components/modals/ConfirmDialog.vue'
import CreateUserDialog from '@/components/modals/CreateUserDialog.vue'
import EditUserDrawer from '@/components/modals/EditUserDrawer.vue'
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { useUsers } from '@/composables/useUsers'
import { useAuthStore } from '@/stores/auth.store'
import { getApiErrorMessage } from '@/utils/error'
import { formatDateTime } from '@/utils/formatters'

import type { ManagedUser } from '@/types'

const authStore = useAuthStore()
const {
  activeUsers,
  invitedUsers,
  loading,
  createUser,
  inviteUser,
  updateUser,
  deleteUser,
  deleteInvite,
  busyIds,
} = useUsers()

const showInvite = ref(false)
const showCreateUser = ref(false)
const showEditDrawer = ref(false)
const showConfirmDelete = ref(false)
const selectedUserId = ref<string | null>(null)
const selectedUser = ref<ManagedUser | null>(null)
const deleteMode = ref<'user' | 'invite'>('user')

const totalManagedUsers = computed(() => activeUsers.value.length + invitedUsers.value.length)
const currentUserId = computed(() => authStore.user?.id ?? null)

function requestEdit(user: ManagedUser) {
  selectedUser.value = user
  showEditDrawer.value = true
}

function requestDelete(id: string, mode: 'user' | 'invite') {
  selectedUserId.value = id
  deleteMode.value = mode
  showConfirmDelete.value = true
}

async function handleInvite(
  payload: { email: string; role: 'admin' | 'user' },
  done: (success: boolean) => void,
) {
  try {
    await inviteUser(payload)
    toast.success('Benutzer wurde eingeladen')
    done(true)
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Fehler beim Einladen'))
    done(false)
  }
}

async function handleCreate(
  payload: { email: string; role: 'admin' | 'user'; password: string },
  done: (success: boolean) => void,
) {
  try {
    await createUser(payload)
    toast.success('Benutzer wurde angelegt')
    done(true)
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Fehler beim Anlegen'))
    done(false)
  }
}

async function handleEdit(
  payload: { id: string; email: string; role: 'admin' | 'user'; password?: string },
  done: (success: boolean) => void,
) {
  try {
    await updateUser(payload.id, {
      email: payload.email,
      role: payload.role,
      password: payload.password,
    })
    toast.success('Benutzer wurde aktualisiert')
    selectedUser.value = null
    done(true)
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Fehler beim Speichern'))
    done(false)
  }
}

async function handleDelete(done: (success: boolean) => void) {
  if (!selectedUserId.value) return done(false)
  const id = selectedUserId.value
  busyIds.value.add(id)
  try {
    if (deleteMode.value === 'invite') {
      await deleteInvite(id)
      toast.success('Einladung wurde entfernt')
    } else {
      await deleteUser(id)
      toast.success('Benutzer wurde gelöscht')
    }
    done(true)
  } catch (e) {
    toast.error(getApiErrorMessage(e, 'Fehler beim Löschen'))
    done(false)
  } finally {
    busyIds.value.delete(id)
  }
}
</script>

<template>
  <div>
    <PageHeader
      title="Benutzer"
      subtitle="Aktive Benutzerkonten und ausstehende Einladungen zentral verwalten."
    >
      <template #actions>
        <Button variant="outline" @click="showCreateUser = true">
          <Users class="mr-1 h-4 w-4" />
          Benutzer anlegen
        </Button>
        <Button @click="showInvite = true">
          <UserPlus class="mr-1 h-4 w-4" />
          Benutzer einladen
        </Button>
      </template>
    </PageHeader>

    <Tabs default-value="accounts" class="gap-4">
      <div class="flex items-center justify-between gap-3">
        <TabsList>
          <TabsTrigger value="accounts">Konten ({{ activeUsers.length }})</TabsTrigger>
          <TabsTrigger value="invites">Einladungen ({{ invitedUsers.length }})</TabsTrigger>
        </TabsList>
        <span class="text-muted-foreground text-sm">{{ totalManagedUsers }} Einträge gesamt</span>
      </div>

      <TabsContent value="accounts">
        <div class="rounded-md border">
          <Table class="table-fixed">
            <TableHeader>
              <TableRow>
                <TableHead class="w-[38%]">E-Mail</TableHead>
                <TableHead class="w-[16%]">Rolle</TableHead>
                <TableHead class="w-[16%]">Status</TableHead>
                <TableHead class="w-[20%]">Erstellt am</TableHead>
                <TableHead class="w-[10%] text-right">Aktionen</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <template v-if="loading">
                <TableRow v-for="i in 3" :key="`account-skeleton-${i}`" class="h-13">
                  <TableCell><Skeleton class="h-4 w-40" /></TableCell>
                  <TableCell><Skeleton class="h-5 w-16 rounded-full" /></TableCell>
                  <TableCell><Skeleton class="h-5 w-18 rounded-full" /></TableCell>
                  <TableCell><Skeleton class="h-4 w-32" /></TableCell>
                  <TableCell><Skeleton class="ml-auto h-4 w-10" /></TableCell>
                </TableRow>
              </template>
              <TableEmpty v-else-if="activeUsers.length === 0" :colspan="5">
                Keine aktiven Benutzer vorhanden.
              </TableEmpty>
              <TableRow v-for="user in activeUsers" v-else :key="user.id">
                <TableCell>
                  <div class="font-medium">{{ user.email }}</div>
                  <div v-if="currentUserId === user.id" class="text-muted-foreground mt-1 text-xs">
                    Dein aktueller Account
                  </div>
                </TableCell>
                <TableCell>
                  <Badge :variant="user.role === 'admin' ? 'default' : 'secondary'">
                    {{ user.role === 'admin' ? 'Admin' : 'User' }}
                  </Badge>
                </TableCell>
                <TableCell>
                  <Badge variant="outline" class="border-emerald-500/30 text-emerald-700">
                    Aktiv
                  </Badge>
                </TableCell>
                <TableCell class="text-muted-foreground">
                  {{ formatDateTime(user.createdAt) }}
                </TableCell>
                <TableCell class="text-right">
                  <div class="flex items-center justify-end gap-1">
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger as-child>
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            :disabled="busyIds.has(user.id)"
                            @click="requestEdit(user)"
                          >
                            <Pencil class="h-4 w-4" />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>Benutzer bearbeiten</TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger as-child>
                          <span>
                            <Button
                              variant="ghost"
                              size="icon-sm"
                              class="text-destructive hover:text-destructive"
                              :disabled="currentUserId === user.id || busyIds.has(user.id)"
                              @click="requestDelete(user.id, 'user')"
                            >
                              <Trash2 class="h-4 w-4" />
                            </Button>
                          </span>
                        </TooltipTrigger>
                        <TooltipContent>
                          {{
                            currentUserId === user.id
                              ? 'Der eigene Benutzer kann nicht gelöscht werden'
                              : 'Benutzer löschen'
                          }}
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </div>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
      </TabsContent>

      <TabsContent value="invites">
        <div class="rounded-md border">
          <Table class="table-fixed">
            <TableHeader>
              <TableRow>
                <TableHead class="w-[40%]">E-Mail</TableHead>
                <TableHead class="w-[20%]">Rolle</TableHead>
                <TableHead class="w-[20%]">Status</TableHead>
                <TableHead class="w-[20%] text-right">Aktionen</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <template v-if="loading">
                <TableRow v-for="i in 3" :key="`invite-skeleton-${i}`" class="h-13">
                  <TableCell><Skeleton class="h-4 w-40" /></TableCell>
                  <TableCell><Skeleton class="h-5 w-16 rounded-full" /></TableCell>
                  <TableCell><Skeleton class="h-5 w-20 rounded-full" /></TableCell>
                  <TableCell><Skeleton class="ml-auto h-4 w-8" /></TableCell>
                </TableRow>
              </template>
              <TableEmpty v-else-if="invitedUsers.length === 0" :colspan="4">
                Keine ausstehenden Einladungen vorhanden.
              </TableEmpty>
              <TableRow v-for="user in invitedUsers" v-else :key="user.id">
                <TableCell>
                  <div class="font-medium">{{ user.email }}</div>
                  <div class="text-muted-foreground mt-1 text-xs">
                    Erstellt am {{ formatDateTime(user.createdAt) }}
                  </div>
                </TableCell>
                <TableCell>
                  <Badge :variant="user.role === 'admin' ? 'default' : 'secondary'">
                    {{ user.role === 'admin' ? 'Admin' : 'User' }}
                  </Badge>
                </TableCell>
                <TableCell>
                  <Badge variant="outline" class="border-amber-500/30 text-amber-700">
                    Ausstehend
                  </Badge>
                </TableCell>
                <TableCell class="text-right">
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger as-child>
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          class="text-destructive hover:text-destructive"
                          :disabled="busyIds.has(user.id)"
                          @click="requestDelete(user.id, 'invite')"
                        >
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
        </div>
      </TabsContent>
    </Tabs>

    <CreateUserDialog v-model:open="showCreateUser" @submit="handleCreate" />
    <InviteUserDialog v-model:open="showInvite" @submit="handleInvite" />
    <EditUserDrawer v-model:open="showEditDrawer" :user="selectedUser" @submit="handleEdit" />

    <ConfirmDialog
      v-model:open="showConfirmDelete"
      :title="deleteMode === 'invite' ? 'Einladung entfernen' : 'Benutzer löschen'"
      :description="
        deleteMode === 'invite'
          ? 'Möchtest du diese ausstehende Einladung wirklich entfernen? Der Benutzer kann sich dann nicht mehr registrieren.'
          : 'Möchtest du diesen Benutzer wirklich löschen? Aktive Sessions und zugehörige Referenzen können dadurch beeinflusst werden.'
      "
      :confirm-label="deleteMode === 'invite' ? 'Entfernen' : 'Löschen'"
      @confirm="handleDelete"
    />
  </div>
</template>
