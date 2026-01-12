<template>
  <div
    class="p-4 rounded-4xl bg-gray-500/20 text-xl font-bold text-center select-none cursor-pointer"
    v-show="isFront"
    @click="onClickCard"
  >
    {{ front }}
  </div>
  <div
    class="p-4 rounded-4xl bg-green-500/20 text-sm text-left select-none cursor-pointer"
    v-show="!isFront"
    @click="onClickCard"
    v-html="back"
  />
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const isFront = ref(true)

const { front, back } = defineProps<{
  front: string,
  back: string,
}>()

watch(() => front, () => {
  isFront.value = true
})

const onClickCard = () => {
  isFront.value = !isFront.value
}
</script>