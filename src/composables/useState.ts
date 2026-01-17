const state = {
  tma: window.Telegram?.WebApp?.initData || null,
  token: localStorage.getItem('token'),
  apiUrl: import.meta.env.VITE_API_URL,
  es: null as EventSource | null,
}

export const useState = () => {
  const isTelegramEnv = () => state.tma !== null
  const isLoggedIn = () => state.token !== null
  const getAuthorizationHeader = () => isTelegramEnv() ? `tma ${state.tma}` : `Bearer ${state.token}`
  const getApiUrl = () => state.apiUrl
  const setToken = (token: string) => {
    localStorage.setItem('token', token)
    state.token = token
  }
  const setES = (es: EventSource | null) => {
    state.es = es
  }
  const getES = () => state.es

  return {
    isTelegramEnv,
    isLoggedIn,
    getAuthorizationHeader,
    getApiUrl,
    setToken,
    setES,
    getES,
  }
}