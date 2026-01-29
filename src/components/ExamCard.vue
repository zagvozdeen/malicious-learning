<template>
  <div class="flex flex-col rounded-4xl bg-gray-500/20 backdrop-blur-lg border border-gray-500/20 shadow-lg">
    <div class="text-center uppercase text-sm font-bold py-1 select-none">
      <slot name="header" />
    </div>
    <span class="h-px w-full bg-gray-500/20" />
    <article
      class="text-xl font-bold text-center p-4 select-none"
      v-html="front"
    />
    <span
      class="h-px w-full bg-gray-500/20"
      v-show="isFront"
    />
    <Transition
      name="collapse-custom"
      @before-enter="beforeEnter"
      @enter="enter"
      @after-enter="afterEnter"
      @before-leave="beforeLeave"
      @leave="leave"
    >
      <div
        class="border-y border-gray-500/20"
        v-show="!isFront"
      >
        <slot />
      </div>
    </Transition>
    <button
      type="button"
      class="hover:bg-gray-500/20 cursor-pointer p-4 rounded-b-4xl font-medium bg-gray-500/15"
      @click="onClickShowButton"
    >
      {{ isFront ? 'Посмотреть ответ' : 'Спрятать ответ' }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const isFront = ref(true)

const { front } = defineProps<{
  front: string,
}>()

watch(() => front, () => {
  isFront.value = true
})

const onClickShowButton = () => {
  isFront.value = !isFront.value
}

const beforeEnter = (el: Element) => {
  if (el instanceof HTMLElement) {
    el.style.height = '0'
  }
}

const enter = (el: Element) => {
  if (el instanceof HTMLElement) {
    el.style.height = el.scrollHeight + 'px'
  }
}

const afterEnter = (el: Element) => {
  if (el instanceof HTMLElement) {
    el.style.height = 'auto'
  }
}

const beforeLeave = (el: Element) => {
  if (el instanceof HTMLElement) {
    el.style.height = el.scrollHeight + 'px'
  }
}

const leave = (el: Element) => {
  if (el instanceof HTMLElement) {
    el.style.height = '0'
  }
}
</script>

<style scoped>
.collapse-custom-enter-active,
.collapse-custom-leave-active {
  transition: height 0.25s ease;
  overflow: hidden; /* Hide overflowing content during transition */
}
</style>
