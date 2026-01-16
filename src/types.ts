export interface TestSession {
    id: number
    uuid: string
    user_id: number
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
    group_uuid: string
    count_null: number
    count_remember: number
    count_forget: number
    created_at: string
}

