const state = {
  tma: window.Telegram?.WebApp?.initData || null,
  token: localStorage.getItem('token'),
  apiUrl: import.meta.env.VITE_API_URL,
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

  return {
    isTelegramEnv,
    isLoggedIn,
    getAuthorizationHeader,
    getApiUrl,
    setToken,
  }
}