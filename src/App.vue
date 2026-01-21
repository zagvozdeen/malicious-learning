<template>
  <div
    class="mx-auto px-4"
    :class="{ [state.getRootClasses.value]: true }"
  >
    <NotificationProvider>
      <router-view />
    </NotificationProvider>
  </div>
</template>

<script setup lang="ts">
import NotificationProvider from '@/components/NotificationProvider.vue'
import { useState } from '@/composables/useState.ts'

const state = useState()
</script>