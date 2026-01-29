import { ref } from 'vue'

const state = {
  tma: window.Telegram?.WebApp?.initData || null,
  token: localStorage.getItem('token'),
  apiUrl: import.meta.env.VITE_API_URL,
  es: null as EventSource | null,
  rootClasses: ref<string>('max-w-md'),
}

export type State = ReturnType<typeof useState>

export const useState = () => {
  return {
    isTelegramEnv: () => state.tma !== null,
    isLoggedIn: () => state.token !== null,
    getAuthorizationHeader: () => state.tma !== null ? `tma ${state.tma}` : `Bearer ${state.token}`,
    getApiUrl: () => state.apiUrl,
    setToken: (token: string) => {
      localStorage.setItem('token', token)
      state.token = token
    },
    unsetToken: () => {
      localStorage.removeItem('token')
      state.token = null
    },
    setES: (es: EventSource | null) => {
      state.es = es
    },
    getES: () => state.es,
    getRootClasses: state.rootClasses,
    setRootClasses: (v: string) => {
      state.rootClasses.value = v
    },
  }
}