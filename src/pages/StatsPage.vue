<template>
  <AppLayout class="max-w-md">
    <canvas ref="ctx" />
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, useTemplateRef } from 'vue'
import Chart from 'chart.js/auto'
import { useFetch } from '@/composables/useFetch.ts'
import { format } from 'date-fns'
import { useState } from '@/composables/useState.ts'
import { useRouter } from 'vue-router'
import AppLayout from '@/components/AppLayout.vue'

const ctx = useTemplateRef('ctx')
const fetcher = useFetch()
const state = useState()
const router = useRouter()
let chart: Chart | null = null

onMounted(() => {
  fetcher
    .getTestSessions()
    .then(data => {
      if (data.ok && ctx.value) {
        const summaries = data.data.data
        const byDate = new Map<string, { empty: number; forget: number; remember: number }>()
        for (const ts of summaries) {
          const d = format(new Date(ts.created_at), 'yyyy-MM-dd')
          const entry = byDate.get(d) ?? { empty: 0, forget: 0, remember: 0 }
          entry.empty += ts.count_null
          entry.forget += ts.count_forget
          entry.remember += ts.count_remember
          byDate.set(d, entry)
        }

        const labels = Array.from(byDate.keys()).sort()
        const datasets = [
          {
            label: 'Не отвечено',
            data: labels.map(label => byDate.get(label)?.empty ?? 0),
            backgroundColor: '#6b7280',
          },
          {
            label: 'Забыл',
            data: labels.map(label => byDate.get(label)?.forget ?? 0),
            backgroundColor: '#ef4444',
          },
          {
            label: 'Вспомнил',
            data: labels.map(label => byDate.get(label)?.remember ?? 0),
            backgroundColor: '#22c55e',
          },
        ]

        if (chart) {
          chart.destroy()
        }
        chart = new Chart(ctx.value, {
          type: 'bar',
          data: {
            labels,
            datasets,
          },
          options: {
            responsive: true,
            scales: {
              x: { stacked: true },
              y: { stacked: true, beginAtZero: true },
            },
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
  if (chart) {
    chart.destroy()
    chart = null
  }
  if (state.isTelegramEnv()) {
    window.Telegram.WebApp.BackButton.hide()
  }
})
</script>
