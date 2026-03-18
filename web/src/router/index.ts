import { createRouter, createWebHistory } from "vue-router";
import HomePage from "@/views/HomePage.vue";
import LoginPage from "@/views/LoginPage.vue";
import AdminLayout from "@/layouts/AdminLayout.vue";
import AdminAuditLogView from "@/views/admin/AdminAuditLogView.vue";
import AdminImagesView from "@/views/admin/AdminImagesView.vue";
import AdminSandboxesView from "@/views/admin/AdminSandboxesView.vue";
import { useAuthStore } from "@/stores/auth";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", name: "home", component: HomePage },
    { path: "/login", name: "login", component: LoginPage },
    {
      path: "/admin",
      component: AdminLayout,
      meta: { requiresAuth: true },
      children: [
        { path: "", redirect: "/admin/images" },
        { path: "images", name: "admin-images", component: AdminImagesView, meta: { requiresAuth: true, title: "Images" } },
        { path: "sandboxes", name: "admin-sandboxes", component: AdminSandboxesView, meta: { requiresAuth: true, title: "Sandboxes" } },
        { path: "audit-log", name: "admin-audit-log", component: AdminAuditLogView, meta: { requiresAuth: true, title: "Audit Log" } },
      ],
    },
  ],
  scrollBehavior() {
    return { top: 0 };
  },
});

router.beforeEach(async (to) => {
  const auth = useAuthStore();

  if (!auth.ready) {
    await auth.bootstrap();
  }

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: "login", query: { redirect: to.fullPath } };
  }

  if (to.name === "login" && auth.isAuthenticated) {
    return { name: "admin-images" };
  }

  return true;
});

export default router;
