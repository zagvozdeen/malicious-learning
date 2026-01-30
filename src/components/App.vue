<template>
  <n-config-provider :theme="darkTheme">
    <n-loading-bar-provider>
      <notification-provider>
        <router-view />
      </notification-provider>
    </n-loading-bar-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { darkTheme, NConfigProvider, NLoadingBarProvider } from 'naive-ui'
import NotificationProvider from '@/components/NotificationProvider.vue'
</script>
