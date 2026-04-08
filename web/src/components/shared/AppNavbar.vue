<script setup lang="ts">
import { LogOut } from 'lucide-vue-next'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'

import logo from '@/assets/logo.png'
import UserAvatar from '@/components/shared/UserAvatar.vue'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useAuthStore } from '@/stores/auth.store'

const authStore = useAuthStore()
const { user, isAuthenticated, isAdmin } = storeToRefs(authStore)
const router = useRouter()

const userTabs = [
  { name: 'Sandboxen', to: '/sandboxes' },
  { name: 'Vorlagen', to: '/images' },
]

const adminTabs = [
  { name: 'Benutzer', to: '/admin/users' },
  { name: 'Protokoll', to: '/admin/audit' },
]

async function handleLogout() {
  await authStore.logout()
  router.push('/')
}
</script>

<template>
  <header class="bg-background sticky top-0 z-50 border-b">
    <div class="mx-auto flex h-14 w-full max-w-6xl items-center gap-4 px-6">
      <RouterLink to="/" class="flex items-center gap-2 text-sm font-semibold">
        <img :src="logo" class="h-7" alt="Shopshredder.de Logo" />
        Shopshredder.de
      </RouterLink>

      <nav v-if="isAuthenticated" class="ml-4 flex items-center gap-1">
        <RouterLink
          v-for="tab in userTabs"
          :key="tab.to"
          :to="tab.to"
          class="hover:bg-accent rounded-md px-3 py-1.5 text-sm transition-colors"
          active-class="bg-accent"
        >
          {{ tab.name }}
        </RouterLink>
        <template v-if="isAdmin">
          <RouterLink
            v-for="tab in adminTabs"
            :key="tab.to"
            :to="tab.to"
            class="hover:bg-accent rounded-md px-3 py-1.5 text-sm transition-colors"
            active-class="bg-accent"
          >
            {{ tab.name }}
          </RouterLink>
        </template>
      </nav>

      <div class="ml-auto flex items-center gap-3">
        <template v-if="isAuthenticated && user">
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <Button variant="ghost" class="relative h-8 w-8 rounded-full">
                <UserAvatar :src="user.avatarUrl" :alt="user.email" class="h-8 w-8" />
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
