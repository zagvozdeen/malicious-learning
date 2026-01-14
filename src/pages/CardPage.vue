<template>
  <div class="min-h-dvh w-full flex items-center justify-center py-24">
    <div class="flex flex-col gap-4 w-full">
      <ExamCard
        :front="currentQuestion.question"
        :back="currentQuestion.answer"
      />

      <div class="fixed w-full max-w-md px-4 top-12 left-1/2 -translate-x-1/2">
        <div
          class="h-8 mb-6 grid gap-0.5 bg-gray-500/20 backdrop-blur-lg border border-gray-500/20 shadow-lg py-1 px-2 rounded-full"
          :style="{ 'grid-template-columns': `repeat(${questions.length}, 1fr)` }"
        >
          <div
            v-for="q in questions"
            :key="q.group_uuid"
            class="flex items-end justify-center pb-0.5 rounded"
            :class="{ [UserAnswerStatusColors[q.status]]: true }"
          >
            <span
              class="size-1.5 rounded-full bg-white"
              v-show="currentQuestion.uuid === q.uuid"
            />
          </div>
        </div>

        <div class="flex items-center gap-2 justify-between">
          <router-link
            class="text-2xl font-bold select-none"
            :to="{ name: 'main' }"
          >
            <span class="text-xl"><span class="uppercase">{{ currentQuestion.module_name }}</span> [{{ currentQuestionIndex + 1 }}/{{ questions.length }}]</span>
          </router-link>
          <div class="grid grid-cols-[1fr_min-content_1fr] gap-1 p-1 bg-gray-500/20 backdrop-blur-lg border border-gray-500/20 rounded-full shadow-lg">
            <button
              class="flex flex-col rounded-full py-1 px-3 transition hover:bg-gray-500/25 cursor-pointer text-xs font-bold"
              type="button"
              @click="onClickPrev"
            >
              <i class="bi bi-chevron-left text-lg" />
            </button>
            <span class="w-px bg-gray-500/20" />
            <button
              class="flex flex-col rounded-full py-1 px-3 transition hover:bg-gray-500/25 cursor-pointer text-xs font-bold"
              type="button"
              @click="onClickNext"
            >
              <i class="bi bi-chevron-right text-lg" />
            </button>
          </div>
        </div>
      </div>

      <div class="fixed w-full max-w-md px-4 bottom-4 left-1/2 -translate-x-1/2">
        <div class="grid grid-cols-[1fr_min-content_1fr] gap-1 p-1 bg-gray-500/20 backdrop-blur-lg border border-gray-500/20 rounded-full shadow-lg">
          <button
            class="flex flex-col rounded-full py-1 px-3 transition hover:bg-gray-500/25 cursor-pointer text-xs font-bold"
            @click="onClickRememberButton"
            type="button"
          >
            <i class="bi bi-lightbulb-fill text-sm" />
            <span>Вспомнил</span>
          </button>
          <span class="w-px bg-gray-500/20" />
          <button
            class="flex flex-col rounded-full py-1 px-3 transition hover:bg-gray-500/25 cursor-pointer text-xs font-bold"
            @click="onClickForgetButton"
            type="button"
          >
            <i class="bi bi-heartbreak-fill text-sm" />
            <span>Забыл</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import ExamCard from '@/components/ExamCard.vue'
import { useRoute } from 'vue-router'
import { useFetch } from '@/store.ts'
import { type UserAnswer, UserAnswerStatus, UserAnswerStatusColors } from '@/types.ts'

const route = useRoute()
const fetcher = useFetch()
const currentQuestionIndex = ref(0)
const questions = ref<UserAnswer[]>([])

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
  }
})

const updateUserAnswer = (uuid: string, status: UserAnswerStatus) => {
  const index = questions.value.findIndex(q => q.uuid === uuid)
  if (questions.value[index]) {
    questions.value[index].status = status
  }
}

const onClickRememberButton = () => {
  fetcher
    .updateUserAnswer(currentQuestion.value.uuid, UserAnswerStatus.Remember)
    .then(data => {
      onClickNext()
      updateUserAnswer(data.uuid, data.status)
    })
}

const onClickForgetButton = () => {
  fetcher
    .updateUserAnswer(currentQuestion.value.uuid, UserAnswerStatus.Forgot)
    .then(data => {
      onClickNext()
      updateUserAnswer(data.uuid, data.status)
    })
}

onMounted(() => {
  fetcher
    .getTestSession(route.params.uuid as string)
    .then(data => {
      questions.value = data.data
    })
})
</script>
