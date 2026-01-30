import '@/style.css'
import { createApp } from 'vue'
import { useState } from '@/composables/useState.ts'
import { createRouter, createWebHistory } from 'vue-router'
import App from '@/components/App.vue'
import CardPage from '@/pages/CardPage.vue'
import MainPage from '@/pages/MainPage.vue'
import LoginPage from '@/pages/LoginPage.vue'
import StatsPage from '@/pages/StatsPage.vue'
import CardsPage from '@/pages/CardsPage.vue'
import CreateTestSessionPage from '@/pages/CreateTestSessionPage.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'main', component: MainPage },
    { path: '/login', name: 'login', component: LoginPage },
    { path: '/cards/:uuid', name: 'cards.view', component: CardPage },
    { path: '/stats', name: 'stats', component: StatsPage },
    { path: '/cards', name: 'cards', component: CardsPage },
    { path: '/cards/create', name: 'cards.create', component: CreateTestSessionPage },
  ],
})

const state = useState()

router.beforeEach((to, _, next) => {
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
