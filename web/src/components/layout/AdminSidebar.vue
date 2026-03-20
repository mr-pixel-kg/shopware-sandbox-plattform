<script setup lang="ts">
import { FolderKanban, Logs, PlaySquare, ShieldCheck } from "lucide-vue-next";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { Badge } from "@/components/ui/badge";

type AdminTab = "images" | "sandboxes" | "audit";

const props = defineProps<{
  active: AdminTab;
}>();

const emit = defineEmits<{
  select: [tab: AdminTab];
}>();

const items: Array<{ key: AdminTab; label: string; icon: typeof FolderKanban }> = [
  { key: "images", label: "Images", icon: FolderKanban },
  { key: "sandboxes", label: "Sandboxes", icon: PlaySquare },
  { key: "audit", label: "Audit Log", icon: Logs },
];
</script>

<template>
  <Sidebar collapsible="none" class="self-start">
    <SidebarHeader>
      <div class="flex items-center gap-3 rounded-lg border bg-background px-3 py-3">
        <div class="flex h-9 w-9 items-center justify-center rounded-md bg-primary text-primary-foreground">
          <ShieldCheck class="h-4 w-4" />
        </div>
        <div class="min-w-0">
          <p class="text-sm font-semibold">Admin area</p>
          <p class="text-xs text-muted-foreground">Internal operations</p>
        </div>
      </div>
    </SidebarHeader>

    <SidebarContent>
      <SidebarGroup>
        <SidebarGroupLabel>Navigation</SidebarGroupLabel>
        <SidebarGroupContent>
          <SidebarMenu>
            <SidebarMenuItem v-for="item in items" :key="item.key">
              <SidebarMenuButton
                :is-active="props.active === item.key"
                @click="emit('select', item.key)"
              >
                <component :is="item.icon" />
                <span>{{ item.label }}</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroupContent>
      </SidebarGroup>
    </SidebarContent>

    <SidebarFooter>
      <div class="flex items-center justify-between rounded-lg border bg-muted/30 px-3 py-3">
        <span class="text-xs text-muted-foreground">Backend access</span>
        <Badge variant="secondary">JWT</Badge>
      </div>
    </SidebarFooter>
  </Sidebar>
</template>
