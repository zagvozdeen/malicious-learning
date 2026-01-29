import type {
  Card,
  Course,
  FullUserAnswer, Module,
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

export interface createTestSessionData {
  course_slug: string
  module_ids: number[]
  shuffle: boolean
}

const createTestSession = async (state: State, notify: Notify, data: createTestSessionData) => {
  return fetchJson<TestSession>(notify, `${state.getApiUrl()}/api/test-sessions`, {
    method: 'POST',
    body: JSON.stringify(data),
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

const getModulesByCourseSlug = async (state: State, notify: Notify, slug: string) => {
  return fetchJson<Module[]>(notify, `${state.getApiUrl()}/api/modules?course_slug=${slug}`, {
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
    createTestSession: (data: createTestSessionData) => createTestSession(state, notify, data),
    getTestSession: (uuid: string) => getTestSession(state, notify, uuid),
    getTestSessions: () => getTestSessions(state, notify),
    updateUserAnswer: (uuid: string, status: UserAnswerStatus) => updateUserAnswer(state, notify, uuid, status),
    getAllCards: () => getAllCards(state, notify),
    getAllCourses: () => getAllCourses(state, notify),
    getModulesByCourseSlug: (slug: string) => getModulesByCourseSlug(state, notify, slug),
  }
}
