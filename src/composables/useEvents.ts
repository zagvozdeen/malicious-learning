import { useState } from '@/composables/useState.ts'

export const useEvents = () => {
  const state = useState()

  const getEventSource = () => {
    const params = new URLSearchParams({
      token: state.getAuthorizationHeader(),
    })
    const es = new EventSource(`${state.getApiUrl()}/api/events?${params.toString()}`)
    state.setES(es)
  }

  return {
    getEventSource,
  }
}
