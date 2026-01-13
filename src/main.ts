import { createApp } from 'vue'
import './style.css'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import CardPage from '@/pages/CardPage.vue'
import MainPage from '@/pages/MainPage.vue'
import type {} from "telegram-web-app";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'main', component: MainPage },
    { path: '/card', name: 'card', component: CardPage },
  ],
})

fetch('http://127.0.0.1:8081/api/auth', {
  method: 'POST',
  body: JSON.stringify({
    username: 'root',
    password: 'password',
  }),
})



createApp(App).use(router).mount('#app')
