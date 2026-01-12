<template>
  <div
    class="h-96 w-full perspective-[1000px] cursor-pointer"
    @click="onClickCard"
  >
    <div
      class="relative h-full w-full transition-all duration-500 transform-3d"
      :class="{ 'rotate-y-180': isFront }"
    >
      <div class="absolute inset-0 flex items-center justify-center rounded-4xl bg-gray-500/20 hover:bg-gray-500/25 p-4 text-white backface-hidden">
        <h1 class="text-xl font-bold text-center">
          {{ front }}
        </h1>
      </div>

      <div class="absolute inset-0 flex items-center justify-center rounded-4xl bg-green-500/20 hover:bg-green-500/25 p-4 text-white rotate-y-180 backface-hidden">
        <span
          class="text-base text-center"
          v-html="back"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const isFront = ref(false)

const { front, back } = defineProps<{
  front: string,
  back: string,
}>()

const onClickCard = () => {
  isFront.value = !isFront.value
}
</script>