import {nextTick, onBeforeUnmount, onMounted, ref} from 'vue'

const focusable = 'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'

export function useDialog(close) {
  const dialogRef = ref(null)
  let previousFocus

  function keydown(event) {
    if (event.key === 'Escape') {
      event.preventDefault()
      close()
      return
    }
    if (event.key !== 'Tab' || !dialogRef.value) return
    const items = [...dialogRef.value.querySelectorAll(focusable)].filter(item => !item.hidden)
    if (!items.length) return
    const first = items[0]
    const last = items.at(-1)
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault()
      last.focus()
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault()
      first.focus()
    }
  }

  onMounted(async () => {
    previousFocus = document.activeElement
    document.body.style.overflow = 'hidden'
    window.addEventListener('keydown', keydown)
    await nextTick()
    dialogRef.value?.querySelector('[autofocus], input, button, [href]')?.focus()
  })

  onBeforeUnmount(() => {
    window.removeEventListener('keydown', keydown)
    document.body.style.overflow = ''
    previousFocus?.focus?.()
  })

  return dialogRef
}
