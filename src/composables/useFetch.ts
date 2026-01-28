import type {
  Card,
  Course,
  FullUserAnswer,
  TestSession,
  TestSessionSummary,
  UserAnswer,
  UserAnswerStatus,
} from '@/types.ts'
import { type State, useState } from '@/composables/useState.ts'
import { type Notify, useNotifications } from '@/composables/useNotifications.ts'
import { i18n } from '@/composables/useI18n.ts'

type ApiResult<T> = { ok: true; data: T } | { ok: false }

const fetchJson = async <T>(notify: Notify, input: RequestInfo, init?: RequestInit): Promise<ApiResult<T>> => {
  const res = await fetch(input, init)

  if (!res.ok) {
    const text = (await res.text()).trim()
    notify.error(i18n[text] || text)
    return { ok: false }
  }

  return { ok: true, data: await res.json() }
}

const getToken = async (state: State, notify: Notify, username: string, password: string) => {
  return fetchJson<{ token: string }>(notify, `${state.getApiUrl()}/api/auth`, {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  })
}

const createTestSession = async (state: State, notify: Notify, shuffle: boolean, modules: number[]) => {
  const params = new URLSearchParams({
    shuffle: shuffle.toString(),
    modules: modules.join(','),
  })
  return fetchJson<TestSession>(notify, `${state.getApiUrl()}/api/test-sessions?${params.toString()}`, {
    method: 'POST',
    headers: {
      'Authorization': state.getAuthorizationHeader(),
    },
  })
}

const getTestSession = async (state: State, notify: Notify, uuid: string) => {
  return fetchJson<{ test_session: TestSession; user_answers: FullUserAnswer[] }>(notify, `${state.getApiUrl()}/api/test-sessions/${uuid}`, {
    headers: {
      'Authorization': state.getAuthorizationHeader(),
    },
  })
}

const getTestSessions = async (state: State, notify: Notify) => {
  return fetchJson<{ data: TestSessionSummary[] }>(notify, `${state.getApiUrl()}/api/test-sessions`, {
    headers: {
      'Authorization': state.getAuthorizationHeader(),
    },
  })
}

const updateUserAnswer = async (state: State, notify: Notify, uuid: string, status: UserAnswerStatus) => {
  return fetchJson<{ data: UserAnswer; test_session: TestSession }>(notify, `${state.getApiUrl()}/api/user-answers/${uuid}`, {
    method: 'PATCH',
    headers: {
      'Authorization': state.getAuthorizationHeader(),
    },
    body: JSON.stringify({ status }),
  })
}

const getAllCards = async (state: State, notify: Notify) => {
  return fetchJson<Card[]>(notify, `${state.getApiUrl()}/api/cards`, {
    headers: {
      'Authorization': state.getAuthorizationHeader(),
    },
  })
}

const getAllCourses = async (state: State, notify: Notify) => {
  return fetchJson<Course[]>(notify, `${state.getApiUrl()}/api/courses`, {
    headers: {
      'Authorization': state.getAuthorizationHeader(),
    },
  })
}

export const useFetch = () => {
  const state = useState()
  const notify = useNotifications()

  return {
    getToken: (username: string, password: string) => getToken(state, notify, username, password),
    createTestSession: (shuffle: boolean, modules: number[]) => createTestSession(state, notify, shuffle, modules),
    getTestSession: (uuid: string) => getTestSession(state, notify, uuid),
    getTestSessions: () => getTestSessions(state, notify),
    updateUserAnswer: (uuid: string, status: UserAnswerStatus) => updateUserAnswer(state, notify, uuid, status),
    getAllCards: () => getAllCards(state, notify),
    getAllCourses: () => getAllCourses(state, notify),
  }
}
