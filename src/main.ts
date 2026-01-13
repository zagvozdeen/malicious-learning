import { createApp } from 'vue'
import './style.css'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import CardPage from '@/pages/CardPage.vue'
import MainPage from '@/pages/MainPage.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'main', component: MainPage },
    { path: '/card', name: 'card', component: CardPage },
  ],
})

createApp(App).use(router).mount('#app')
