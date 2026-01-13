import type { UserAnswer, UserAnswerStatus } from '@/types.ts'

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

export const useFetch = () => {
  const state = useState()

  const getToken = async (username: string, password: string) => {
    const res = await fetch(`${state.getApiUrl()}/api/auth`, {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    })
    return await res.json() as { token: string }
  }

  const createTestSession = async (shuffle: boolean, modules: number[]) => {
    const params = new URLSearchParams({
      shuffle: shuffle.toString(),
      modules: modules.join(','),
    })
    const res = await fetch(`${state.getApiUrl()}/api/test-sessions?${params.toString()}`, {
      method: 'POST',
      headers: {
        'Authorization': state.getAuthorizationHeader(),
      },
    })
    return await res.json() as { group_uuid: string }
  }

  const getTestSession = async (uuid: string) => {
    const res = await fetch(`${state.getApiUrl()}/api/test-sessions/${uuid}`, {
      headers: {
        'Authorization': state.getAuthorizationHeader(),
      },
    })
    return await res.json() as { data: UserAnswer[] }
  }

  const updateUserAnswer = async (uuid: string, status: UserAnswerStatus) => {
    const res = await fetch(`${state.getApiUrl()}/api/user-answers/${uuid}`, {
      method: 'PATCH',
      headers: {
        'Authorization': state.getAuthorizationHeader(),
      },
      body: JSON.stringify({ status }),
    })
    return await res.json() as { data: UserAnswer[] }
  }

  return {
    getToken,
    createTestSession,
    getTestSession,
    updateUserAnswer,
  }
}
