<script lang="ts" setup>
import SidebarNav from './components/layout/SidebarNav.vue'
</script>

<template>
  <div class="flex h-screen bg-mist-50">
    <SidebarNav />
    <main class="flex-1 ml-64 overflow-y-auto">
      <router-view />
    </main>
  </div>
</template>
