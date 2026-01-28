<template>
  <div
    ref="swipeDiv"
    class="min-h-dvh w-full flex items-center justify-center"
    :class="{
      'pb-34': ts && ts.is_active,
      'pb-12': !ts || (ts && !ts.is_active),
    }"
    style="padding-top: calc(var(--tg-content-safe-area-inset-top, calc(var(--spacing) * 12)) + var(--tg-safe-area-inset-top, 0px))"
  >
    <div class="flex flex-col gap-4 w-full">
      <AppSpinner v-if="loading" />
      <ExamCard
        v-if="!loading && ts && ts.is_active"
        :front="currentQuestion.question"
        :back="currentQuestion.answer"
      >
        <span>{{ currentQuestion.module_name }} [{{ currentQuestionIndex + 1 }}/{{ questions.length }}]</span>&nbsp;<span
          v-show="currentQuestion.status == 'forgot'"
          class="text-red-500"
        >[вы забыли]</span><span
          v-show="currentQuestion.status == 'remember'"
          class="text-green-500"
        >[вы вспомнили]</span>
      </ExamCard>

      <div
        v-if="ts && !ts.is_active"
        class="flex flex-col rounded-4xl bg-gray-500/20 backdrop-blur-lg border border-gray-500/20 shadow-lg"
      >
        <div class="text-center uppercase text-sm font-bold py-1 select-none">
          Результаты
        </div>
        <span class="h-px w-full bg-gray-500/20" />
        <div
          v-if="loadingResults"
          class="flex flex-col gap-4 my-2 p-4"
        >
          <span class="text-center font-bold">Рекомендации рассчитываются</span>
          <AppSpinner />
        </div>
        <div
          v-else
          class="p-4"
        >
          <div
            v-if="ts.recommendations"
            v-html="ts.recommendations"
            class="text-justify"
          />
          <div
            v-else
            class="text-center font-medium"
          >
            Не удалось получить рекомендации, пожалуйста, попробуйте начать новый тест чтобы получить рекомендации
          </div>
        </div>
        <span class="h-px w-full bg-gray-500/20" />
        <div class="grid sm:grid-cols-10 grid-cols-5 gap-2 p-4">
          <div
            v-for="q in questions"
            :key="q.id"
            class="rounded text-center font-bold"
            :class="{ [UserAnswerStatusColors[q.status]]: true }"
          >
            {{ q.uid }}
          </div>
        </div>
        <router-link
          :to="{ name: 'main' }"
          class="hover:bg-gray-500/20 cursor-pointer p-4 rounded-b-4xl font-medium bg-gray-500/15 text-center"
        >
          На главную
        </router-link>
      </div>

      <div
        v-if="ts && ts.is_active"
        class="fixed flex flex-col gap-2 w-full max-w-md px-4 bottom-4 left-1/2 -translate-x-1/2"
      >
        <div
          v-if="!loading"
          class="h-8 grid gap-0.5 bg-gray-500/20 backdrop-blur-lg border border-gray-500/20 shadow-lg py-1 px-2 rounded-full"
          :style="{ 'grid-template-columns': `repeat(${questions.length}, 1fr)` }"
        >
          <div
            v-for="q in questions"
            :key="q.id"
            class="flex items-end justify-center pb-0.5 rounded"
            :class="{ [UserAnswerStatusColors[q.status]]: true }"
          >
            <span
              class="size-1.5 rounded-full bg-white"
              v-show="currentQuestion.uuid === q.uuid"
            />
          </div>
        </div>

        <div class="grid grid-cols-[min-content_1fr_min-content] gap-2">
          <button
            class="flex items-center justify-center py-1 px-4.5 transition bg-gray-500/20 backdrop-blur-lg border border-gray-500/20 rounded-full shadow-lg hover:bg-gray-500/25 cursor-pointer"
            type="button"
            @click="onClickPrev"
          >
            <i class="bi bi-chevron-left text-base flex" />
          </button>
          <div class="grid grid-cols-[1fr_min-content_1fr] gap-1 p-1 bg-gray-500/20 backdrop-blur-lg border border-gray-500/20 rounded-full shadow-lg">
            <button
              class="flex flex-col rounded-full py-1 px-3 transition hover:bg-gray-500/25 cursor-pointer text-xs font-bold text-center"
              @click="onClickRememberButton"
              type="button"
            >
              <i class="bi bi-lightbulb-fill text-sm" />
              <span>Вспомнил</span>
            </button>
            <span class="w-px bg-gray-500/20" />
            <button
              class="flex flex-col rounded-full py-1 px-3 transition hover:bg-gray-500/25 cursor-pointer text-xs font-bold text-center"
              @click="onClickForgetButton"
              type="button"
            >
              <i class="bi bi-heartbreak-fill text-sm" />
              <span>Забыл</span>
            </button>
          </div>
          <button
            class="flex items-center justify-center py-1 px-4.5 transition bg-gray-500/20 backdrop-blur-lg border border-gray-500/20 rounded-full shadow-lg hover:bg-gray-500/25 cursor-pointer"
            type="button"
            @click="onClickNext"
          >
            <i class="bi bi-chevron-right text-base flex" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, useTemplateRef } from 'vue'
import ExamCard from '@/components/ExamCard.vue'
import { useRoute, useRouter } from 'vue-router'
import { useFetch } from '@/composables/useFetch.ts'
import { useState } from '@/composables/useState.ts'
import {
  type FullUserAnswer,
  type TestSession,
  UserAnswerStatus,
  UserAnswerStatusColors,
} from '@/types.ts'
import AppSpinner from '@/components/AppSpinner.vue'
import { useNotifications } from '@/composables/useNotifications.ts'

const route = useRoute()
const router = useRouter()
const state = useState()
const fetcher = useFetch()
const notify = useNotifications()
const currentQuestionIndex = ref(0)
const loading = ref(true)
const loadingResults = ref(false)
const ts = ref<TestSession | null>(null)
const questions = ref<FullUserAnswer[]>([])
const swipeDiv = useTemplateRef('swipeDiv')
let touchstartX = 0
let touchendX = 0

const onClickPrev = () => {
  if (currentQuestionIndex.value > 0) {
    currentQuestionIndex.value -= 1
  }
}

const onClickNext = () => {
  if (currentQuestionIndex.value < questions.value.length - 1) {
    currentQuestionIndex.value += 1
  }
}

const currentQuestion = computed(() => {
  return questions.value[currentQuestionIndex.value] ?? {
    uuid: '',
    question: 'Вопрос не найден',
    answer: 'Ответ не найден',
    module_name: 'Модуль',
    status: UserAnswerStatus.Null,
    updated_at: Date.now(),
  }
})

const updateUserAnswer = (uuid: string, status: UserAnswerStatus) => {
  const index = questions.value.findIndex(q => q.uuid === uuid)
  if (questions.value[index]) {
    questions.value[index].status = status
  }
}

const updateUserAnswerStatus = (uuid: string, status: UserAnswerStatus) => {
  if (loading.value) {
    notify.warn('Данные ещё загружаются, подождите, пожалуйста')
    return
  }
  fetcher
    .updateUserAnswer(uuid, status)
    .then(data => {
      if (data.ok) {
        onClickNext()
        updateUserAnswer(data.data.data.uuid, data.data.data.status)

        if (ts.value) {
          ts.value.is_active = data.data.test_session.is_active
          ts.value.recommendations = data.data.test_session.recommendations
          ts.value.updated_at = data.data.test_session.updated_at

          if (!ts.value.is_active) {
            notify.info('Вы успешно прошли весь тест, поздравляю!')
          }
        }
      }
    })
}

const onClickRememberButton = () => {
  updateUserAnswerStatus(currentQuestion.value.uuid, UserAnswerStatus.Remember)
}

const onClickForgetButton = () => {
  updateUserAnswerStatus(currentQuestion.value.uuid, UserAnswerStatus.Forgot)
}

const handleGesture = () => {
  const swipeDistance = touchendX - touchstartX

  if (Math.abs(swipeDistance) > 100) {
    if (swipeDistance > 0) {
      onClickPrev()
    } else {
      onClickNext()
    }
  }
}

const handleRecommendationsEnd = (msg: MessageEvent) => {
  loadingResults.value = false
  if (ts.value) {
    ts.value.recommendations = msg.data
  }
}

const handleRecommendationsStart = (msg: MessageEvent) => {
  loadingResults.value = true
  if (ts.value) {
    ts.value.recommendations = msg.data
  }
}

onMounted(() => {
  fetcher
    .getTestSession(route.params.uuid as string)
    .then(data => {
      if (data.ok) {
        ts.value = data.data.test_session
        questions.value = data.data.user_answers

        const i = questions.value.findIndex(q => q.status === UserAnswerStatus.Null)
        if (i !== -1) {
          currentQuestionIndex.value = i
        }

        loading.value = false
      }
    })

  if (swipeDiv.value) {
    swipeDiv.value.addEventListener('touchstart', (e) => {
      touchstartX = e.changedTouches[0]?.screenX || 0
    })
    swipeDiv.value.addEventListener('touchend', (e) => {
      touchendX = e.changedTouches[0]?.screenX || 0
      handleGesture()
    })
  }

  if (state.isTelegramEnv()) {
    window.Telegram.WebApp.BackButton.show()
    window.Telegram.WebApp.BackButton.onClick(() => {
      router.push({ name: 'main' })
    })
  }

  state.getES()?.addEventListener('get-recommendations-start', handleRecommendationsStart)
  state.getES()?.addEventListener('get-recommendations-end', handleRecommendationsEnd)
})

onUnmounted(() => {
  if (state.isTelegramEnv()) {
    window.Telegram.WebApp.BackButton.hide()
  }

  state.getES()?.removeEventListener('get-recommendations-start', handleRecommendationsStart)
  state.getES()?.removeEventListener('get-recommendations-end', handleRecommendationsEnd)
})
</script>
