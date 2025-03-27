import { createRouter, createWebHistory } from "vue-router";

import StartView from "./views/StartView.vue";
import LoginView from "./views/LoginView.vue";
import AdminView from "./views/AdminView.vue";
import AdminSandboxEnvironments from "./views/admin/AdminSandboxEnvironmentsView.vue";
import AdminSandboxImages from "./views/admin/AdminSandboxImagesView.vue";
import ApiService from "@/services/apiService.js";
import AdminAuditLogView from "@/views/admin/AdminAuditLogView.vue";

const routes = [
  { path: "/", component: StartView },
  { path: "/login", component: LoginView },
  {
    path: "/admin",
    component: AdminView,
    redirect: "/admin/sandbox-environments",
    meta: { requiresAuth: true },
    children: [
      { path: "sandbox-environments", component: AdminSandboxEnvironments },
      { path: "sandbox-images", component: AdminSandboxImages },
      { path: "auditlog", component: AdminAuditLogView },
    ],
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach(async (to, from, next) => {
  // Check each route for the meta field 'requiresAuth'
  if (to.matched.some((record) => record.meta.requiresAuth)) {
    const isLoggedIn = await ApiService.isLoggedIn();

    // Go to login page if not logged in
    if (!isLoggedIn) {
      return next("/login");
    }
  }
  // Allow navigation
  next();
});

export default router;
