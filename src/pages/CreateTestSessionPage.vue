<template>
  <div class="min-h-dvh w-full flex items-center justify-center">
    <n-form
      @submit.prevent="onSubmitForm"
      class="w-full"
    >
      <n-form-item
        label="Курс"
        path="course_slug"
      >
        <n-select
          v-model:value="form.course_slug"
          placeholder="Выберите курс"
          :options="courseOptions"
        />
      </n-form-item>
      <n-form-item
        v-show="modules.length > 0"
        label="Модули"
        path="module_ids"
      >
        <n-checkbox-group v-model:value="form.module_ids">
          <n-space vertical>
            <n-checkbox
              v-for="module in modules"
              :key="module.id"
              :value="module.id"
            >
              {{ module.name }}
            </n-checkbox>
          </n-space>
        </n-checkbox-group>
      </n-form-item>
      <n-form-item label="Перемешать">
        <n-switch v-model:value="form.shuffle" />
      </n-form-item>
      <n-form-item
        :show-feedback="false"
        :show-label="false"
      >
        <n-button
          attr-type="submit"
          type="success"
        >
          Начать
        </n-button>
      </n-form-item>
    </n-form>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { type createTestSessionData, useFetch } from '@/composables/useFetch.ts'
import { useState } from '@/composables/useState.ts'
import { NForm, NFormItem, NSelect, NCheckboxGroup, NCheckbox, NSpace, NSwitch, NButton, type SelectOption } from 'naive-ui'
import type { Module } from '@/types.ts'
import { useNotifications } from '@/composables/useNotifications.ts'

const router = useRouter()
const state = useState()
const fetcher = useFetch()
const notify = useNotifications()
const form = reactive({
  course_slug: null as string | null,
  module_ids: [] as number[],
  shuffle: false as boolean,
})
const courseOptions = ref<SelectOption[]>([])
const modules = ref<Module[]>([])

const onSubmitForm = () => {
  if (!form.course_slug) {
    notify.error('Чтобы начать курс, нужно выбрать курс')
    return
  }
  if (form.module_ids.length === 0) {
    notify.error('Чтобы начать тест, нужно выбрать хотя бы один модуль')
    return
  }

  const payload: createTestSessionData = {
    course_slug: form.course_slug,
    module_ids: form.module_ids,
    shuffle: form.shuffle,
  }

  fetcher
    .createTestSession(payload)
    .then(data => {
      if (data.ok) {
        router.push({
          name: 'cards.view',
          params: {
            uuid: data.data.uuid,
          },
        })
        notify.info('Тест начат, вы можете начать прохождение!')
      }
    })
}

watch(() => form.course_slug, (slug) => {
  if (slug) {
    fetcher
      .getModulesByCourseSlug(slug)
      .then(data => {
        if (data.ok) {
          modules.value = data.data
          form.module_ids = modules.value.map((module) => module.id)
        }
      })
  }
})

onMounted(() => {
  fetcher
    .getAllCourses()
    .then(data => {
      if (data.ok) {
        courseOptions.value = data.data.map((course) => ({
          value: course.slug,
          label: course.name,
        }))
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
