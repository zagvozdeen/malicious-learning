<template>
  <teleport to="body">
    <transition-group
      name="notification"
      tag="div"
      class="max-w-md w-full flex flex-col gap-2 px-4 pt-2 fixed left-1/2 -translate-x-1/2"
      style="top: var(--tg-content-safe-area-inset-top, 0px);"
      appear
    >
      <div
        class="bg-gray-500/20 backdrop-blur-lg border border-gray-500/20 shadow-lg py-2 px-3 rounded-xl grid grid-cols-[min-content_1fr] items-center gap-2 select-none"
        v-for="n in notifications"
        :key="n.id"
      >
        <i
          class="bi"
          :class="{
            'bi-info-circle-fill text-blue-400': n.level === 'info',
            'bi-exclamation-circle-fill text-orange-400': n.level === 'warn',
            'bi-x-circle-fill text-red-400': n.level === 'error',
          }"
        />
        <span class="font-medium text-sm">{{ n.msg }}</span>
      </div>
    </transition-group>
  </teleport>

  <slot />
</template>

<script setup lang="ts">
import { provide, ref } from 'vue'
import type { Levels, PusherFunc, Notification } from '@/types.ts'

const notifications = ref<Notification[]>([])

let counter = 0
const pusher: PusherFunc = (level: Levels, msg: string, date: number) => {
  const id = counter++
  notifications.value.unshift({
    id: id,
    level: level,
    msg: msg,
    date: date,
  } as Notification)

  setTimeout(() => {
    const index = notifications.value.findIndex((item) => item.id === id)
    if (index !== -1) {
      notifications.value.splice(index, 1)
    }
  }, 5000)
}

provide('notifications', pusher)
</script>

<style scoped>
.notification-enter-active,
.notification-leave-active,
.notification-move {
  transition: transform 0.35s ease, opacity 0.35s ease;
}

.notification-enter-from {
  opacity: 0;
  transform: translateY(-18px);
}

.notification-enter-to {
  opacity: 1;
  transform: translateY(0);
}

.notification-leave-from {
  opacity: 1;
  transform: translateX(0);
}

.notification-leave-to {
  opacity: 0;
  transform: translateX(60px);
}
</style>
