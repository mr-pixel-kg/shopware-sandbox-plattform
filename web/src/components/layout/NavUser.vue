<script setup lang="ts">
import { computed } from "vue";
import { ChevronsUpDown, LogOut } from "lucide-vue-next";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar";

const props = defineProps<{
  user: {
    name: string;
    email: string;
    avatar: string;
  };
}>();

const emit = defineEmits<{
  logout: [];
}>();

const { isMobile } = useSidebar();

const fallback = computed(() => {
  const source = props.user.name || props.user.email;
  return source.slice(0, 2).toUpperCase();
});
</script>

<template>
  <SidebarMenu>
    <SidebarMenuItem>
      <DropdownMenu>
        <DropdownMenuTrigger as-child>
          <SidebarMenuButton
            size="lg"
            class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
          >
            <Avatar size="sm" shape="square">
              <AvatarImage :src="user.avatar" :alt="user.name" />
              <AvatarFallback shape="square">
                {{ fallback }}
              </AvatarFallback>
            </Avatar>
            <div class="grid flex-1 text-left text-sm leading-tight">
              <span class="truncate font-medium">{{ user.name }}</span>
              <span class="truncate text-xs">{{ user.email }}</span>
            </div>
            <ChevronsUpDown class="ml-auto size-4" />
          </SidebarMenuButton>
        </DropdownMenuTrigger>

        <DropdownMenuContent
          class="w-[--reka-dropdown-menu-trigger-width] min-w-56 rounded-lg"
          :side="isMobile ? 'bottom' : 'right'"
          align="start"
          :side-offset="4"
        >
          <DropdownMenuLabel class="p-0 font-normal">
            <div class="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
              <Avatar size="sm" shape="square">
                <AvatarImage :src="user.avatar" :alt="user.name" />
                <AvatarFallback shape="square">
                  {{ fallback }}
                </AvatarFallback>
              </Avatar>
              <div class="grid flex-1 text-left text-sm leading-tight">
                <span class="truncate font-medium">{{ user.name }}</span>
                <span class="truncate text-xs">{{ user.email }}</span>
              </div>
            </div>
          </DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem @click="emit('logout')">
            <LogOut />
            Log out
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </SidebarMenuItem>
  </SidebarMenu>
</template>
