<template>
  <div class="min-h-dvh w-full flex items-center justify-center">
    <div class="w-full flex flex-col items-center gap-4">
      <div
        class="group h-64 w-64 perspective-[1000px] cursor-pointer"
        @click="test = !test"
      >
        <div
          class="relative h-full w-full transition-all duration-500 transform-3d group-hosver:transform-[rotateY(180deg)]"
          :class="{ 'transform-[rotateY(180deg)]': test }"
        >
          <div class="absolute inset-0 flex items-center justify-center rounded bg-blue-500 text-white backface-hidden">
            <h1 class="text-xl font-bold">
              Привет! (Лицо)
            </h1>
          </div>

          <div class="absolute inset-0 flex items-center justify-center rounded bg-red-500 text-white transform-[rotateY(180deg)] backface-hidden">
            <h1 class="text-xl font-bold">
              Пока! (Оборот)
            </h1>
          </div>
        </div>
      </div>

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
