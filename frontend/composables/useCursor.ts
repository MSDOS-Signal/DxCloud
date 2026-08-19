import { onBeforeUnmount, onMounted } from 'vue'

/**
 * Moonshot 风格自定义光标：外圈延迟跟随 + 内点紧跟，系统指针隐藏。
 * 触屏设备自动禁用；输入框等交互元素恢复系统光标保证可用性。
 */
export function useCursor() {
  let dot: HTMLDivElement | null = null
  let ring: HTMLDivElement | null = null
  let raf = 0
  let mx = 0
  let my = 0
  let rx = 0
  let ry = 0
  let visible = false

  function loop() {
    rx += (mx - rx) * 0.18
    ry += (my - ry) * 0.18
    if (ring) ring.style.transform = `translate(${rx}px, ${ry}px)`
    if (dot) dot.style.transform = `translate(${mx}px, ${my}px)`
    raf = requestAnimationFrame(loop)
  }

  function onMove(e: MouseEvent) {
    mx = e.clientX
    my = e.clientY
    if (!visible) {
      visible = true
      dot?.classList.add('dx-cursor-visible')
      ring?.classList.add('dx-cursor-visible')
    }
    const t = e.target as HTMLElement
    const interactive = t.closest(
      'input, textarea, select, [contenteditable="true"], .n-input, .n-base-selection, [role="combobox"], a, button, .n-button',
    )
    if (ring) {
      if (interactive) ring.classList.add('dx-cursor-hover')
      else ring.classList.remove('dx-cursor-hover')
    }
  }

  function onDown() {
    ring?.classList.add('dx-cursor-down')
  }

  function onUp() {
    ring?.classList.remove('dx-cursor-down')
  }

  function onLeave() {
    visible = false
    dot?.classList.remove('dx-cursor-visible')
    ring?.classList.remove('dx-cursor-visible')
  }

  onMounted(() => {
    if (window.matchMedia('(hover: none)').matches) return
    dot = document.createElement('div')
    dot.className = 'dx-cursor-dot'
    ring = document.createElement('div')
    ring.className = 'dx-cursor-ring'
    document.body.appendChild(dot)
    document.body.appendChild(ring)
    document.documentElement.classList.add('dx-cursor-on')
    window.addEventListener('mousemove', onMove, { passive: true })
    window.addEventListener('mousedown', onDown)
    window.addEventListener('mouseup', onUp)
    document.addEventListener('mouseleave', onLeave)
    raf = requestAnimationFrame(loop)
  })

  onBeforeUnmount(() => {
    cancelAnimationFrame(raf)
    window.removeEventListener('mousemove', onMove)
    window.removeEventListener('mousedown', onDown)
    window.removeEventListener('mouseup', onUp)
    document.removeEventListener('mouseleave', onLeave)
    dot?.remove()
    ring?.remove()
    document.documentElement.classList.remove('dx-cursor-on')
  })
}
