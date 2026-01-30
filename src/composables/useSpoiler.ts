export const onSpoilerContainerClick = (e: unknown): void => {
  if (e instanceof Event) {
    if (e.target instanceof Element) {
      const el = e.target.closest('.spoiler')
      if (el && !el.classList.contains('_show')) {
        el.classList.add('_show')
      }
    }
  }
}