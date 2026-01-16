import { inject } from 'vue'
import type { PusherFunc } from '@/types.ts'

export const useNotifications = () => {
  const pusher = inject('notifications') as PusherFunc

  return {
    info: (n: string) => pusher('info', n, Date.now()),
    warn: (n: string) => pusher('warn', n, Date.now()),
    error: (n: string) => pusher('error', n, Date.now()),
  }
}