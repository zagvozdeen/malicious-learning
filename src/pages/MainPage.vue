<template>
  <AppLayout class="max-w-md">
    <div class="min-h-dvh w-full flex flex-col gap-4 items-center justify-center py-6">
      <ul class="flex flex-col gap-px w-full rounded-2xl border border-gray-500/30 overflow-hidden">
        <li class="w-full">
          <router-link
            class="grid grid-cols-[min-content_1fr_min-content] items-center w-full gap-2 p-2 cursor-pointer bg-gray-500/20 hover:bg-gray-500/30"
            type="button"
            :to="{ name: 'cards' }"
          >
            <span class="size-6 flex items-center justify-center rounded-lg bg-gray-400">
              <i class="bi bi-list-check text-sm flex" />
            </span>
            <span class="text-left text-sm font-medium">Все карточки</span>
            <span class="text-gray-400">
              <i class="bi bi-chevron-right text-sm flex" />
            </span>
          </router-link>
        </li>
        <li class="w-full">
          <router-link
            class="grid grid-cols-[min-content_1fr_min-content] items-center w-full gap-2 p-2 cursor-pointer bg-gray-500/20 hover:bg-gray-500/30"
            type="button"
            :to="{ name: 'stats' }"
          >
            <span class="size-6 flex items-center justify-center rounded-lg bg-orange-400">
              <i class="bi bi-graph-up-arrow text-sm flex" />
            </span>
            <span class="text-left text-sm font-medium">Статистика</span>
            <span class="text-gray-400">
              <i class="bi bi-chevron-right text-sm flex" />
            </span>
          </router-link>
        </li>
        <li class="w-full">
          <router-link
            class="grid grid-cols-[min-content_1fr_min-content] items-center w-full gap-2 p-2 cursor-pointer bg-gray-500/20 hover:bg-gray-500/30"
            type="button"
            :to="{ name: 'cards.create' }"
          >
            <span class="size-6 flex items-center justify-center rounded-lg bg-purple-400">
              <i class="bi bi-bookmark-star-fill text-sm flex" />
            </span>
            <span class="text-left text-sm font-medium">Начать тест</span>
            <span class="text-gray-400">
              <i class="bi bi-chevron-right text-sm flex" />
            </span>
          </router-link>
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
            :to="{ name: 'cards.view', params: { uuid: ts.uuid } }"
          >
            <span class="size-6 flex items-center justify-center rounded-lg bg-gray-500">
              <span class="text-gray-200 font-bold">{{ index + 1 }}</span>
            </span>
            <div class="flex flex-col">
              <div class="flex items-center gap-1">
                <span class="text-left text-sm font-medium">Тест по курсу «{{ ts.course_name }}» от {{ format(ts.created_at, "dd.MM.yyyy HH:mm") }}</span>
                <i
                  v-if="!ts.is_active"
                  class="bi bi-check-all text-lg flex"
                />
              </div>
              <span class="text-gray-400 text-xs">{{ ts.is_shuffled ? 'Вопросы вперемешку' : 'Вопросы по порядку' }}, {{ pluralize(ts.module_ids.length, ['выбран', 'выбрано', 'выбрано']) }} {{ ts.module_ids.length }} {{ pluralize(ts.module_ids.length, ['модуль', 'модуля', 'модулей']) }}</span>
            </div>
            <span class="flex items-center gap-1 text-gray-400">
              <AppPercent :ts="ts" />
              <i class="bi bi-chevron-right text-sm flex" />
            </span>
          </router-link>
        </li>
      </ul>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { useFetch } from '@/composables/useFetch.ts'
import { onMounted, ref } from 'vue'
import type { TestSessionSummary } from '@/types.ts'
import { format } from 'date-fns'
import AppPercent from '@/components/AppPercent.vue'
import { pluralize } from '@/composables/useI18n.ts'
import AppLayout from '@/components/AppLayout.vue'

const fetcher = useFetch()
const testSessions = ref<TestSessionSummary[]>([])

onMounted(() => {
  fetcher
    .getTestSessions()
    .then(data => {
      if (data.ok) {
        testSessions.value = data.data.data
      }
    })
})
</script>
