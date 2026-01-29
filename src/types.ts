export interface TestSession {
    id: number
    uuid: string
    user_id: number
    course_id: number
    module_ids: number[]
    is_shuffled: boolean
    is_active: boolean
    recommendations: string | null
    created_at: string
    updated_at: string
}

export interface UserAnswer {
    id: number
    uuid: string
    card_id: number
    test_session_id: number
    status: UserAnswerStatus
    created_at: string
    updated_at: string
}

export interface FullUserAnswer {
    id: number
    uid: number
    uuid: string
    card_id: number
    test_session_id: number
    status: UserAnswerStatus
    created_at: string
    updated_at: string
    answer: string
    question: string
    module_id: number
    module_name: string
}

export const UserAnswerStatus = {
  Null: 'null',
  Remember: 'remember',
  Forgot: 'forgot',
} as const

export type UserAnswerStatus = (typeof UserAnswerStatus)[keyof typeof UserAnswerStatus];

export const UserAnswerStatusColors: Record<UserAnswerStatus, string> = {
  [UserAnswerStatus.Null]: 'bg-gray-500',
  [UserAnswerStatus.Remember]: 'bg-green-500',
  [UserAnswerStatus.Forgot]: 'bg-red-500',
} as const

export interface TestSessionSummary {
    uuid: string
    is_active: boolean
    is_shuffled: boolean
    module_ids: number[]
    has_recommendations: boolean
    count_null: number
    count_remember: number
    count_forget: number
    created_at: string
    course_name: string
}

export type Levels = 'error'| 'warn' | 'info'

export interface Notification {
    id: number
    level: Levels
    msg: string
    date: number
}

export type PusherFunc = (level: Levels, msg: string, date: number) => void

export interface Card {
    id: number
    uid: number
    uuid: string
    question: string
    answer: string
    module_id: number
    is_active: boolean
    hash: string
    created_at: string
    updated_at: string
}

export interface Course {
    id: number
    uuid: string
    slug: string
    name: string
    updated_at: string
    created_at: string
}

export interface Module {
    id: number
    uuid: string
    name: string
    updated_at: string
    created_at: string
}
