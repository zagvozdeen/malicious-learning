import '@/style.css'
import { createApp, h } from 'vue'
import { useState } from '@/composables/useState.ts'
import { createRouter, createWebHistory, RouterView } from 'vue-router'
import { useEvents } from '@/composables/useEvents.ts'
import CardPage from '@/pages/CardPage.vue'
import MainPage from '@/pages/MainPage.vue'
import LoginPage from '@/pages/LoginPage.vue'
import StatsPage from '@/pages/StatsPage.vue'
import CardsPage from '@/pages/CardsPage.vue'
import NotificationProvider from '@/components/NotificationProvider.vue'
import CreateTestSessionPage from '@/pages/CreateTestSessionPage.vue'
import { darkTheme, NConfigProvider } from 'naive-ui'

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

createApp({
  setup: () => () => h('div', {
    class: `mx-auto px-4 ${state.getRootClasses.value}`,
  }, h(NConfigProvider, {
    theme: darkTheme,
  }, () => h(NotificationProvider, () => h(RouterView)))),
}).use(router).mount('#app')
