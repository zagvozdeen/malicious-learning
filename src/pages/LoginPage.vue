<template>
  <div class="min-h-dvh w-full flex items-center justify-center">
    <form @submit.prevent="onSubmitForm">
      <input
        type="text"
        placeholder="Username"
        v-model="form.username"
      >
      <input
        type="password"
        placeholder="Password"
        v-model="form.password"
      >
      <button type="submit">
        Login
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { reactive } from 'vue'
import { useFetch } from '@/composables/useFetch.ts'
import { useState } from '@/composables/useState.ts'

const router = useRouter()
const state = useState()
const fetcher = useFetch()
const form = reactive({
  username: '',
  password: '',
})

const onSubmitForm = () => {
  fetcher
    .getToken(form.username, form.password)
    .then(data => {
      if (data) {
        state.setToken(data.token)
        router.push({ name: 'main' })
      }
    })
}
</script>
