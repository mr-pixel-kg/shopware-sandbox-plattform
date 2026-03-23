import { createRouter, createWebHistory } from 'vue-router'

import { useAuthStore } from '@/stores/auth.store'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/auth/LoginView.vue'),
      meta: { layout: 'auth', guest: true },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/auth/RegisterView.vue'),
      meta: { layout: 'auth', guest: true },
    },
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/explore/ExploreView.vue'),
    },
    {
      path: '/sandboxes',
      name: 'sandboxes',
      component: () => import('@/views/sandboxes/SandboxesView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/images',
      name: 'images',
      component: () => import('@/views/admin/AdminImagesView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/admin/users',
      name: 'admin-users',
      component: () => import('@/views/admin/AdminUsersView.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
    },
    {
      path: '/admin/audit',
      name: 'admin-audit',
      component: () => import('@/views/admin/AdminAuditView.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/',
    },
  ],
})

router.beforeEach((to) => {
  const authStore = useAuthStore()

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  if (to.meta.guest && authStore.isAuthenticated) {
    return { name: 'sandboxes' }
  }

  if (to.meta.requiresAdmin && !authStore.isAdmin) {
    return { name: 'sandboxes' }
  }
})

export default router
