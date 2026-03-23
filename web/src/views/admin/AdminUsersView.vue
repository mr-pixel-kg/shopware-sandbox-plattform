<script setup lang="ts">
import { UserPlus } from 'lucide-vue-next'
import { ref } from 'vue'

import InviteUserDialog from '@/components/modals/InviteUserDialog.vue'
import EmptyState from '@/components/shared/EmptyState.vue'
import PageHeader from '@/components/shared/PageHeader.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'

const showInvite = ref(false)
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

    <Card>
      <CardContent class="pt-6">
        <EmptyState
          title="Benutzerverwaltung wird bald verfügbar"
          description="Die API für Benutzerverwaltung ist noch in Entwicklung."
        />
      </CardContent>
    </Card>

    <InviteUserDialog v-model:open="showInvite" />
  </div>
</template>
