import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
  static targets = ["fgColor", "fgHex", "bgColor", "bgHex", "preview", "ratio", "check"]

  connect() {
    const savedFg = localStorage.getItem('wcag-contrast-fg')
    const savedBg = localStorage.getItem('wcag-contrast-bg')

    if (savedFg) {
      this.fgHexTarget.value = savedFg
      this.fgColorTarget.value = savedFg
    }
    if (savedBg) {
      this.bgHexTarget.value = savedBg
      this.bgColorTarget.value = savedBg
    }

    this.update()
  }

  // Normalize a user-entered hex string to "#rrggbb", or null if invalid.
  // Accepts "#rgb", "rgb", "#rrggbb", "rrggbb" in any case.
  normalizeHex(value) {
    let v = value.trim().replace(/^#/, '')
    if (/^[0-9a-fA-F]{3}$/.test(v)) {
      v = v.split('').map(c => c + c).join('')
    }
    if (/^[0-9a-fA-F]{6}$/.test(v)) {
      return '#' + v.toLowerCase()
    }
    return null
  }

  // WCAG relative luminance for a normalized "#rrggbb" color.
  luminance(hex) {
    const channel = (i) => {
      const c = parseInt(hex.slice(i, i + 2), 16) / 255
      return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4)
    }
    return 0.2126 * channel(1) + 0.7152 * channel(3) + 0.0722 * channel(5)
  }

  // WCAG contrast ratio between two normalized colors (1 to 21).
  contrastRatio(fg, bg) {
    const l1 = this.luminance(fg)
    const l2 = this.luminance(bg)
    const lighter = Math.max(l1, l2)
    const darker = Math.min(l1, l2)
    return (lighter + 0.05) / (darker + 0.05)
  }

  update() {
    const fg = this.normalizeHex(this.fgHexTarget.value)
    const bg = this.normalizeHex(this.bgHexTarget.value)

    this.fgHexTarget.classList.toggle('is-invalid', !fg)
    this.bgHexTarget.classList.toggle('is-invalid', !bg)

    if (!fg || !bg) {
      this.ratioTarget.textContent = '—'
      this.checkTargets.forEach(check => {
        const badge = check.querySelector('.check-badge')
        badge.textContent = '—'
        badge.classList.remove('pass', 'fail')
      })
      return
    }

    localStorage.setItem('wcag-contrast-fg', fg)
    localStorage.setItem('wcag-contrast-bg', bg)

    const ratio = this.contrastRatio(fg, bg)
    this.ratioTarget.textContent = ratio.toFixed(2) + ':1'

    this.previewTarget.style.setProperty('--preview-fg', fg)
    this.previewTarget.style.setProperty('--preview-bg', bg)

    this.checkTargets.forEach(check => {
      const threshold = parseFloat(check.dataset.threshold)
      const pass = ratio >= threshold
      const badge = check.querySelector('.check-badge')
      badge.textContent = pass ? 'Pass' : 'Fail'
      badge.classList.toggle('pass', pass)
      badge.classList.toggle('fail', !pass)
    })
  }

  // Actions
  fgColorChanged() {
    this.fgHexTarget.value = this.fgColorTarget.value
    this.update()
  }

  fgHexChanged() {
    const hex = this.normalizeHex(this.fgHexTarget.value)
    if (hex) this.fgColorTarget.value = hex
    this.update()
  }

  bgColorChanged() {
    this.bgHexTarget.value = this.bgColorTarget.value
    this.update()
  }

  bgHexChanged() {
    const hex = this.normalizeHex(this.bgHexTarget.value)
    if (hex) this.bgColorTarget.value = hex
    this.update()
  }

  swap() {
    const fg = this.fgHexTarget.value
    this.fgHexTarget.value = this.bgHexTarget.value
    this.bgHexTarget.value = fg

    const fgHex = this.normalizeHex(this.fgHexTarget.value)
    const bgHex = this.normalizeHex(this.bgHexTarget.value)
    if (fgHex) this.fgColorTarget.value = fgHex
    if (bgHex) this.bgColorTarget.value = bgHex

    this.update()
  }
}
