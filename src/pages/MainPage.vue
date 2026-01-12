<template>
  <div class="min-h-dvh w-full flex items-center justify-center">
    <div class="w-full flex flex-col items-center gap-4">
      <button
        class="bg-green-700 hover:bg-green-800 transition uppercase font-bold text-sm w-full rounded p-2 cursor-pointer"
        type="button"
        @click="onAllQuestions"
      >
        Все вопросы подряд
      </button>
      <button
        class="bg-green-700 hover:bg-green-800 transition uppercase font-bold text-sm w-full rounded p-2 cursor-pointer"
        type="button"
        @click="onAllQuestions"
      >
        Все вопросы вперемешку
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { ref } from 'vue'

const test = ref(false)
const router = useRouter()
const onAllQuestions = () => {
  router.push({ name: 'card' })
}
</script>
