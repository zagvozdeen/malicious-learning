<template>
  <div class="min-h-dvh w-full flex flex-col gap-4 items-center justify-center py-6">
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
      <li class="w-full">
        <button
          class="grid grid-cols-[min-content_1fr_min-content] items-center w-full gap-2 p-2 cursor-pointer bg-gray-500/20 hover:bg-gray-500/30"
          type="button"
          @click="onOnlyTheoryQuestions"
        >
          <span class="size-6 flex items-center justify-center rounded-lg bg-orange-400">
            <i class="bi bi-question-square-fill text-sm flex" />
          </span>
          <span class="text-left text-sm font-medium">Только теория</span>
          <span class="text-gray-400">
            <i class="bi bi-chevron-right text-sm flex" />
          </span>
        </button>
      </li>
      <li class="w-full">
        <button
          class="grid grid-cols-[min-content_1fr_min-content] items-center w-full gap-2 p-2 cursor-pointer bg-gray-500/20 hover:bg-gray-500/30"
          type="button"
          @click="onOnlyPracticeQuestions"
        >
          <span class="size-6 flex items-center justify-center rounded-lg bg-amber-400">
            <i class="bi bi-keyboard-fill text-sm flex" />
          </span>
          <span class="text-left text-sm font-medium">Только практика</span>
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
        v-for="(ts, index) in testSessions"
        :key="ts.uuid"
      >
        <router-link
          class="grid grid-cols-[min-content_1fr_min-content] items-center w-full gap-2 p-2 cursor-pointer bg-gray-500/20 hover:bg-gray-500/30"
          type="button"
          :to="{ name: 'cards', params: { uuid: ts.uuid } }"
        >
          <span class="size-6 flex items-center justify-center rounded-lg bg-gray-500">
            <span class="text-gray-200 font-bold">{{ index + 1 }}</span>
          </span>
          <div class="flex flex-col">
            <div class="flex items-center gap-1">
              <span class="text-left text-sm font-medium">Тест от {{ format(ts.created_at, "dd.MM.yyyy HH:mm:ss") }}</span>
              <i
                v-if="!ts.is_active"
                class="bi bi-check-all text-lg flex"
              />
            </div>
            <span class="text-gray-400 text-xs">{{ ts.is_shuffled ? 'Вопросы вперемешку' : 'Вопросы по порядку' }}, {{ ts.module_ids.length == 1 ? (ts.module_ids[0] == 1 ? 'только теория' : 'только практика') : 'теория и практика' }}</span>
          </div>
          <span class="flex items-center gap-1 text-gray-400">
            <AppPercent :ts="ts" />
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
import AppPercent from '@/components/AppPercent.vue'

const router = useRouter()
const fetcher = useFetch()
const notify = useNotifications()
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
        notify.info('Тест начат, вы можете начать прохождение!')
      }
    })
}

const onAllQuestions = () => {
  createTestSession(false, [1, 2])
}

const onAllShuffleQuestions = () => {
  createTestSession(true, [1, 2])
}

const onOnlyTheoryQuestions = () => {
  createTestSession(false, [1])
}

const onOnlyPracticeQuestions = () => {
  createTestSession(false, [2])
}

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
