<template>
  <teleport to="body">
    <div class="max-w-md w-full flex flex-col gap-2 px-4 pt-2 fixed top-0 left-1/2 -translate-x-1/2">
      <div
        class="bg-gray-500/20 backdrop-blur-lg border border-gray-500/20 shadow-lg py-2 px-3 rounded-xl grid grid-cols-[min-content_1fr] items-center gap-2"
        v-for="n in notifications"
        :key="n.id"
      >
        <i
          class="bi "
          :class="{
            'bi-info-circle-fill text-blue-400': n.level === 'info',
            'bi-exclamation-circle-fill text-orange-400': n.level === 'warn',
            'bi-x-circle-fill text-red-400': n.level === 'error',
          }"
        />
        <span class="font-medium text-sm">{{ n.msg }}</span>
      </div>
    </div>
  </teleport>

  <slot />
</template>

<script setup lang="ts">
import { provide, ref } from 'vue'
import type { Levels, PusherFunc, Notification } from '@/types.ts'

const notifications = ref<Notification[]>([])

let counter = 0
const pusher: PusherFunc = (level: Levels, msg: string, date: number) => {
  notifications.value.push({
    id: counter++,
    level: level,
    msg: msg,
    date: date,
  } as Notification)
}

provide('notifications', pusher)
</script>

<style scoped>

</style>