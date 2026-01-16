import type { FullUserAnswer, TestSession, TestSessionSummary, UserAnswer, UserAnswerStatus } from '@/types.ts'
import { useState } from '@/composables/useState.ts'
import { useNotifications } from '@/composables/useNotifications.ts'
import { i18n } from '@/composables/useI18n.ts'

export const useFetch = () => {
  const state = useState()
  const notify = useNotifications()

  const handleError = async (res: Response) => {
    if (!res.ok) {
      const text = (await res.text()).trim()
      notify.error(i18n[text] || text)
      return true
    }
    return false
  }

  const getToken = async (username: string, password: string) => {
    const res = await fetch(`${state.getApiUrl()}/api/auth`, {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    })
    if (await handleError(res)) return
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
    if (await handleError(res)) return
    return await res.json() as TestSession
  }

  const getTestSession = async (uuid: string) => {
    const res = await fetch(`${state.getApiUrl()}/api/test-sessions/${uuid}`, {
      headers: {
        'Authorization': state.getAuthorizationHeader(),
      },
    })
    if (await handleError(res)) return
    return await res.json() as { test_session: TestSession; user_answers: FullUserAnswer[] }
  }

  const getTestSessions = async () => {
    const res = await fetch(`${state.getApiUrl()}/api/test-sessions`, {
      headers: {
        'Authorization': state.getAuthorizationHeader(),
      },
    })
    if (await handleError(res)) return
    return await res.json() as { data: TestSessionSummary[] }
  }

  const updateUserAnswer = async (uuid: string, status: UserAnswerStatus) => {
    const res = await fetch(`${state.getApiUrl()}/api/user-answers/${uuid}`, {
      method: 'PATCH',
      headers: {
        'Authorization': state.getAuthorizationHeader(),
      },
      body: JSON.stringify({ status }),
    })
    if (await handleError(res)) return
    return await res.json() as { data: UserAnswer; test_session: TestSession }
  }

  return {
    getToken,
    createTestSession,
    getTestSession,
    getTestSessions,
    updateUserAnswer,
  }
}
