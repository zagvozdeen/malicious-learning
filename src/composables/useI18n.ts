export const i18n = {
  'user answer status must be null': 'Вы уже ответили на этот вопрос, если хотите ответить на вопрос повторно, то начните новый тест',
  'test session is not active': 'Этот тест устарел и закрыт, начните новый тест',
} as Readonly<Record<string, string>>

const cases = [2, 0, 1, 1, 1, 2]

export const pluralize = (count: number, words: string[]): string => {
  return words[ (count % 100 > 4 && count % 100 < 20) ? 2 : cases[Math.min(count % 10, 5)] as number ]
}
