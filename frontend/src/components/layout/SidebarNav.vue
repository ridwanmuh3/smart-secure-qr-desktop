<script lang="ts" setup>
  import { useRoute } from 'vue-router'
  import { Icon } from '@iconify/vue'

  const route = useRoute()

  const navItems = [
    { path: '/issuer', label: 'Buat QR Aman', icon: 'lucide:file-text' },
    { path: '/verifier', label: 'Verifikasi QR', icon: 'lucide:shield-check' },
    { path: '/keys', label: 'Kelola Kunci', icon: 'lucide:key-round' },
  ]
</script>

<template>
  <nav class="fixed left-0 top-0 h-screen w-64 bg-emerald-900 text-white flex flex-col shadow-xl z-10">
    <div class="p-5 border-b border-emerald-700">
      <h1 class="text-lg font-bold tracking-tight">Smart Secure QR-Code</h1>
    </div>
    <div class="flex-1 py-4">
      <router-link v-for="item in navItems" :key="item.path" :to="item.path"
        class="flex items-center gap-3 px-5 py-3 mx-2 rounded-lg text-sm font-medium transition-all duration-200"
        :class="route.path === item.path
          ? 'bg-emerald-700 text-white shadow-md'
          : 'text-emerald-200 hover:bg-emerald-800 hover:text-white'">
        <Icon :icon="item.icon" class="w-5 h-5 shrink-0" />
        <span>{{ item.label }}</span>
      </router-link>
    </div>
    <div class="p-4 border-t border-emerald-700">
      <p class="text-emerald-300 text-xs font-semibold text-center">&copy; 2025 FAST Unsil. All Rights Reserved.</p>
    </div>
  </nav>
</template>
