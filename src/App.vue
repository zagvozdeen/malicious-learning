<template>
  <div class="max-w-md mx-auto px-4">
test {{ test }} test
    <router-view />
  </div>
</template>

<script setup lang="ts">
const test = window.Telegram?.WebApp?.initData || 'nothing'
</script>