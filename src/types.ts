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

