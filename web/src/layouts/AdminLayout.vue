<script setup lang="ts">
import { computed } from "vue";
import { useRouter, RouterLink, RouterView } from "vue-router";
import { Boxes, FolderKanban, Logs, PlaySquare } from "lucide-vue-next";
import { useAuthStore } from "@/stores/auth";
import NavUser from "@/components/layout/NavUser.vue";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarRail,
  SidebarSeparator,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { Separator } from "@/components/ui/separator";

const auth = useAuthStore();
const router = useRouter();

const navUser = computed(() => {
  const email = auth.user?.email ?? "";
  const name = email.split("@")[0] || "User";

  return {
    name,
    email,
    avatar: "",
  };
});

const items = [
  { title: "Images", to: "/admin/images", icon: FolderKanban },
  { title: "Sandboxes", to: "/admin/sandboxes", icon: PlaySquare },
  { title: "Audit Log", to: "/admin/audit-log", icon: Logs },
];

function logout() {
  auth.logout();
  router.push("/login");
}
</script>

<template>
  <SidebarProvider>
    <Sidebar variant="inset">
      <SidebarHeader>
        <div class="flex items-center gap-3 px-2 py-2">
          <div class="flex h-8 w-8 items-center justify-center rounded-md bg-primary text-primary-foreground">
            <Boxes class="h-4 w-4" />
          </div>
          <div class="min-w-0">
            <div class="truncate text-sm font-semibold">Shopshredder</div>
            <div class="text-xs text-muted-foreground">Sandbox Platform</div>
          </div>
        </div>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Platform</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem v-for="item in items" :key="item.to">
                <SidebarMenuButton as-child>
                  <RouterLink :to="item.to">
                    <component :is="item.icon" />
                    <span>{{ item.title }}</span>
                  </RouterLink>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter>
        <NavUser :user="navUser" @logout="logout" />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>

    <SidebarInset>
      <header class="flex h-14 shrink-0 items-center gap-2 bg-background px-4">
        <SidebarTrigger class="-ml-1" />
        <Separator orientation="vertical" class="mr-2 h-4" />
        <div class="text-sm font-medium">Admin</div>
      </header>

      <main class="flex-1 bg-muted/30 p-4 md:p-6">
        <RouterView />
      </main>
    </SidebarInset>
  </SidebarProvider>
</template>
