<template>
  <canvas ref="ctx" />
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, useTemplateRef } from 'vue'
import Chart from 'chart.js/auto'
import { useFetch } from '@/composables/useFetch.ts'
import { format } from 'date-fns'
import { useState } from '@/composables/useState.ts'
import { useRouter } from 'vue-router'

const ctx = useTemplateRef('ctx')
const fetcher = useFetch()
const state = useState()
const router = useRouter()

onMounted(() => {
  fetcher
    .getTestSessions()
    .then(data => {
      if (data.ok && ctx.value) {
        const label = ['2026-01-16', '2026-01-17']
        const values = [
          { label: 'null', data: [0, 0] },
          { label: 'forget', data: [0, 0] },
          { label: 'remember', data: [0, 0] },
        ]
        for (const ts of data.data.data) {
          const d = format(ts.created_at, 'yyyy-MM-dd')
          const i = label.findIndex(v => v === d)
          if (values[0] && values[0].data[i] !== undefined) values[0].data[i] += ts.count_null
          if (values[1] && values[1].data[i] !== undefined) values[1].data[i] += ts.count_forget
          if (values[2] && values[2].data[i] !== undefined) values[2].data[i] += ts.count_remember
        }

        new Chart(ctx.value, {
          type: 'bar',
          data: {
            labels: label,
            datasets: values,
          },
        })
      }
    })

  if (state.isTelegramEnv()) {
    window.Telegram.WebApp.BackButton.show()
    window.Telegram.WebApp.BackButton.onClick(() => {
      router.push({ name: 'main' })
    })
  }
})

onUnmounted(() => {
  if (state.isTelegramEnv()) {
    window.Telegram.WebApp.BackButton.hide()
  }
})
</script>
