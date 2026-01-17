<template>
  <span
    class="text-xs text-green-400 font-semibold"
    :class="{
      'text-red-400': percent >= 0 && percent < 25,
      'text-yellow-400': percent >= 25 && percent < 50,
      'text-orange-400': percent >= 50 && percent < 75,
      'text-green-400': percent >= 75 && percent <= 100,
    }"
  >{{ percent }}%</span>
</template>

<script lang="ts" setup>
import type { TestSessionSummary } from '@/types.ts'

const { ts } = defineProps<{
  ts: TestSessionSummary
}>()

const percent = Math.round(((ts.count_forget + ts.count_remember) / (ts.count_remember + ts.count_forget + ts.count_null)) * 10000) / 100
</script>