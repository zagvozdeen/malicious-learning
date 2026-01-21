<template>
  <div class="pdf flex flex-col gap-6 py-6">
    <h2 class="text-xl font-medium text-center">
      Содержание
    </h2>
    <ol class="flex flex-col list-decimal list-inside gap-1 text-xs">
      <li
        v-for="card in cards"
        :key="card.uid"
        :value="card.uid"
        class="truncate max-w-full"
      >
        <h3 class="inline">
          <a
            :href="`#u${card.uid}`"
            class="hover:underline *:inline"
            v-html="card.question"
          />
        </h3>
      </li>
    </ol>
    <h2 class="text-xl font-medium text-center mt-6">
      Ответы
    </h2>
    <ul class="flex flex-col gap-6">
      <li
        :id="`u${card.uid}`"
        v-for="card in cards"
        :key="card.uid"
      >
        <h3 class="mb-2">
          <a
            :href="`#u${card.uid}`"
            class="hover:underlinef hover:text-gray-300"
            v-html="card.question"
          />
        </h3>
        <article
          class="text-sm text-justify flex flex-col gap-2 *:list-inside border-l-3 pl-4 py-1 rounded-l"
          v-html="card.answer"
        />
        <span class="block h-px w-full bg-gray-400/50 mt-6" />
      </li>
    </ul>
  </div>
</template>

<script lang="ts" setup>
import { useFetch } from '@/composables/useFetch.ts'
import { onMounted, ref } from 'vue'
import type { Card } from '@/types.ts'
import { useState } from '@/composables/useState.ts'

const state = useState()
const fetcher = useFetch()
const cards = ref<Card[]>([])

onMounted(() => {
  state.setRootClasses('max-w-2xl')

  fetcher
    .getAllCards()
    .then(data => {
      if (data) {
        cards.value = data
      }
    })
})
</script>