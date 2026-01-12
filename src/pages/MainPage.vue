<template>
  <div class="min-h-dvh w-full flex items-center justify-center">
    <ul class="flex flex-col gap-px w-full rounded-2xl border border-gray-500/30 overflow-hidden">
      <li class="w-full">
        <button
          class="grid grid-cols-[min-content_1fr_min-content] items-center w-full gap-2 p-2 cursor-pointer bg-gray-500/20 hover:bg-gray-500/30"
          type="button"
          @click="onAllQuestions"
        >
          <span class="size-6 flex items-center justify-center rounded-lg bg-blue-400">
            <i class="bi bi-check-square-fill text-sm" />
          </span>
          <span class="text-left text-sm font-medium">Все вопросы подряд</span>
          <span class="text-gray-400">
            <i class="bi bi-chevron-right text-sm" />
          </span>
        </button>
      </li>
      <li class="w-full">
        <button
          class="grid grid-cols-[min-content_1fr_min-content] items-center w-full gap-2 p-2 cursor-pointer bg-gray-500/20 hover:bg-gray-500/30"
          type="button"
          @click="onAllQuestions"
        >
          <span class="size-6 flex items-center justify-center rounded-lg bg-red-400">
            <i class="bi bi-shuffle text-sm" />
          </span>
          <span class="text-left text-sm font-medium">Все вопросы вперемешку</span>
          <span class="text-gray-400">
            <i class="bi bi-chevron-right text-sm" />
          </span>
        </button>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()
const onAllQuestions = () => {
  router.push({ name: 'card' })
}
</script>
