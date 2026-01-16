<template>
  <div class="min-h-dvh w-full flex flex-col gap-4 items-center justify-center">
    <ul class="flex flex-col gap-px w-full rounded-2xl border border-gray-500/30 overflow-hidden">
      <li class="w-full">
        <button
          class="grid grid-cols-[min-content_1fr_min-content] items-center w-full gap-2 p-2 cursor-pointer bg-gray-500/20 hover:bg-gray-500/30"
          type="button"
          @click="onAllQuestions"
        >
          <span class="size-6 flex items-center justify-center rounded-lg bg-blue-400">
            <i class="bi bi-check-square-fill text-sm flex" />
          </span>
          <span class="text-left text-sm font-medium">Все вопросы подряд</span>
          <span class="text-gray-400">
            <i class="bi bi-chevron-right text-sm flex" />
          </span>
        </button>
      </li>
      <li class="w-full">
        <button
          class="grid grid-cols-[min-content_1fr_min-content] items-center w-full gap-2 p-2 cursor-pointer bg-gray-500/20 hover:bg-gray-500/30"
          type="button"
          @click="onAllShuffleQuestions"
        >
          <span class="size-6 flex items-center justify-center rounded-lg bg-red-400">
            <i class="bi bi-shuffle text-sm flex" />
          </span>
          <span class="text-left text-sm font-medium">Все вопросы вперемешку</span>
          <span class="text-gray-400">
            <i class="bi bi-chevron-right text-sm flex" />
          </span>
        </button>
      </li>
    </ul>

    <ul
      v-if="testSessions.length > 0"
      class="flex flex-col gap-px w-full rounded-2xl border border-gray-500/30 overflow-hidden"
    >
      <li
        class="w-full"
        v-for="(testSession, index) in testSessions"
        :key="testSession.group_uuid"
      >
        <router-link
          class="grid grid-cols-[min-content_1fr_min-content] items-center w-full gap-2 p-2 cursor-pointer bg-gray-500/20 hover:bg-gray-500/30"
          type="button"
          :to="{ name: 'cards', params: { uuid: testSession.group_uuid } }"
        >
          <span
            class="size-6 flex items-center justify-center rounded-lg"
            :class="{[colors[index % 5] as string]: true}"
          >
            <i
              class="bi text-sm flex"
              :class="{[`bi-${index % 10 + 1}-circle-fill`]: true}"
            />
          </span>
          <span class="text-left text-sm font-medium">Тест от {{ format(testSession.created_at, "dd.MM.yyyy HH:mm:ss") }}</span>
          <span class="flex items-center gap-1 text-gray-400">
            <span class="font-semibold"><span class="text-xs text-green-400 font-semibold">{{ testSession.count_remember }}</span>/<span class="text-xs text-red-400">{{ testSession.count_forget }}</span></span>
            <i class="bi bi-chevron-right text-sm flex" />
          </span>
        </router-link>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useFetch } from '@/composables/useFetch.ts'
import { onMounted, ref } from 'vue'
import type { TestSessionSummary } from '@/types.ts'
import { format } from 'date-fns'
import { useNotifications } from '@/composables/useNotifications.ts'

const router = useRouter()
const fetcher = useFetch()
const notifications = useNotifications()
const testSessions = ref<TestSessionSummary[]>([])

const createTestSession = (shuffle: boolean, modules: number[]) => {
  fetcher
    .createTestSession(shuffle, modules)
    .then(data => {
      if (data) {
        router.push({
          name: 'cards',
          params: {
            uuid: data.uuid,
          },
        })
      }
    })
}

const onAllQuestions = () => {
  createTestSession(false, [1, 2])
}

const onAllShuffleQuestions = () => {
  createTestSession(true, [1, 2])
}

const colors = ['bg-green-400', 'bg-yellow-400', 'bg-blue-400', 'bg-red-400', 'bg-orange-400']

onMounted(() => {
  fetcher
    .getTestSessions()
    .then(data => {
      if (data) {
        testSessions.value = data.data
      }
    })
})
</script>
