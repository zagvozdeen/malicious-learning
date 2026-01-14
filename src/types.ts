export interface UserAnswer {
    uuid: string
    group_uuid: string
    card_id: number
    status: UserAnswerStatus
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

