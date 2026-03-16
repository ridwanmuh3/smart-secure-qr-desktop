import { createRouter, createWebHashHistory } from 'vue-router'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', redirect: '/issuer' },
    {
      path: '/issuer',
      name: 'Issuer',
      component: () => import('../views/IssuerView.vue'),
    },
    {
      path: '/verifier',
      name: 'Verifier',
      component: () => import('../views/VerifierView.vue'),
    },
    {
      path: '/keys',
      name: 'KeyManagement',
      component: () => import('../views/KeyManagementView.vue'),
    },
  ],
})

export default router
