import { createApp } from 'vue'
import './style.css'
import { useState } from '@/store.ts'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import CardPage from '@/pages/CardPage.vue'
import MainPage from '@/pages/MainPage.vue'
import LoginPage from '@/pages/LoginPage.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'main', component: MainPage },
    { path: '/login', name: 'login', component: LoginPage },
    { path: '/card', name: 'card', component: CardPage },
  ],
})

const state = useState()

router.beforeEach((to, _from, next) => {
  if (state.isTelegramEnv()) {
    next()
  } else if (to.name !== 'login' && !state.isLoggedIn()) {
    next({ name: 'login' })
  } else if (to.name === 'login' && state.isLoggedIn()) {
    next({ name: 'main' })
  } else {
    next()
  }
})

createApp(App).use(router).mount('#app')
