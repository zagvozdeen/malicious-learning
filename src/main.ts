import '@/style.css'
import { createApp } from 'vue'
import { useState } from '@/composables/useState.ts'
import { createRouter, createWebHistory } from 'vue-router'
import { useEvents } from '@/composables/useEvents.ts'
import App from '@/App.vue'
import CardPage from '@/pages/CardPage.vue'
import MainPage from '@/pages/MainPage.vue'
import LoginPage from '@/pages/LoginPage.vue'
import StatsPage from '@/pages/StatsPage.vue'
import CardsPage from '@/pages/CardsPage.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'main', component: MainPage },
    { path: '/login', name: 'login', component: LoginPage },
    { path: '/cards/:uuid', name: 'cards', component: CardPage },
    { path: '/stats', name: 'stats', component: StatsPage },
    { path: '/cards', name: 'pdf', component: CardsPage },
  ],
})

const state = useState()

useEvents().getEventSource()

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
