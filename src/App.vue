<template>
  <div class="max-w-md mx-auto px-4">
    <NotificationProvider>
      <router-view />
    </NotificationProvider>
  </div>
</template>

<script setup lang="ts">
import NotificationProvider from '@/components/NotificationProvider.vue'
</script>