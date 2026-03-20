<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useAuthStore } from '@/stores/auth.store'
import { useRouter } from 'vue-router'
import { Separator } from '@/components/ui/separator'
import { Button } from '@/components/ui/button'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { LogOut } from 'lucide-vue-next'
import logo from '@/assets/logo.png'

const authStore = useAuthStore()
const { user, isAuthenticated, isAdmin } = storeToRefs(authStore)
const router = useRouter()

const userTabs = [
  { name: 'Sandboxes', to: '/sandboxes' },
  { name: 'Entdecken', to: '/explore' },
]

const adminTabs = [
  { name: 'Instanzen', to: '/admin/instances' },
  { name: 'Vorlagen', to: '/admin/images' },
  { name: 'Benutzer', to: '/admin/users' },
  { name: 'Protokoll', to: '/admin/audit' },
]

function getInitials(email: string): string {
  const parts = email.split('@')[0].split(/[._-]/)
  return parts
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase() ?? '')
    .join('')
}

async function handleLogout() {
  await authStore.logout()
  router.push('/explore')
}
</script>

<template>
  <header class="sticky top-0 z-50 border-b bg-background">
    <div class="mx-auto w-full max-w-6xl flex h-14 items-center gap-4 px-6">
      <RouterLink to="/" class="flex items-center gap-2 text-sm font-semibold">
        <img :src="logo" class="h-7" alt="Shopshredder.de Logo" />
        Shopshredder.de
      </RouterLink>

      <nav v-if="isAuthenticated" class="flex items-center gap-1 ml-4">
        <RouterLink
          v-for="tab in userTabs"
          :key="tab.to"
          :to="tab.to"
          class="px-3 py-1.5 text-sm rounded-md transition-colors hover:bg-accent"
          active-class="bg-accent"
        >
          {{ tab.name }}
        </RouterLink>
      </nav>
      <RouterLink
        v-else
        to="/explore"
        class="ml-4 px-3 py-1.5 text-sm rounded-md transition-colors hover:bg-accent"
        active-class="bg-accent"
      >
        Entdecken
      </RouterLink>

      <template v-if="isAdmin">
        <Separator orientation="vertical" class="h-6" />
        <nav class="flex items-center gap-1">
          <RouterLink
            v-for="tab in adminTabs"
            :key="tab.to"
            :to="tab.to"
            class="px-3 py-1.5 text-sm rounded-md transition-colors hover:bg-accent"
            active-class="bg-accent"
          >
            {{ tab.name }}
          </RouterLink>
        </nav>
      </template>

      <div class="ml-auto flex items-center gap-3">
        <template v-if="isAuthenticated && user">
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <Button variant="ghost" class="relative h-8 w-8 rounded-full">
                <Avatar class="h-8 w-8">
                  <AvatarFallback>{{ getInitials(user.email) }}</AvatarFallback>
                </Avatar>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" class="w-56">
              <DropdownMenuLabel>
                <div class="flex flex-col space-y-1">
                  <p class="text-sm font-medium">{{ user.email }}</p>
                </div>
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem @click="handleLogout">
                <LogOut class="mr-2 h-4 w-4" />
                Abmelden
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </template>
        <template v-else>
          <Button variant="ghost" size="sm" as-child>
            <RouterLink to="/login">Anmelden</RouterLink>
          </Button>
        </template>
      </div>
    </div>
  </header>
</template>
