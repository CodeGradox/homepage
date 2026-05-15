import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
  static targets = ["sentinel", "headerFrozen", "headerScroll", "headerTable", "bodyScroll"]

  connect() {
    this.handleBodyScroll = this.handleBodyScroll.bind(this)
    this.handleIntersection = this.handleIntersection.bind(this)

    this.bodyScrollTarget.addEventListener("scroll", this.handleBodyScroll)

    this.syncHeaderWidth()
    this.syncHorizontalScroll(this.bodyScrollTarget.scrollLeft)

    this.resizeObserver = new ResizeObserver(() => {
      this.syncHeaderWidth()
      this.syncHorizontalScroll(this.bodyScrollTarget.scrollLeft)
    })

    this.resizeObserver.observe(this.bodyScrollTarget)

    this.intersectionObserver = new IntersectionObserver(this.handleIntersection, {
      threshold: 0,
      rootMargin: "0px"
    })

    this.intersectionObserver.observe(this.sentinelTarget)
  }

  disconnect() {
    this.bodyScrollTarget.removeEventListener("scroll", this.handleBodyScroll)

    if (this.resizeObserver) {
      this.resizeObserver.disconnect()
    }

    if (this.intersectionObserver) {
      this.intersectionObserver.disconnect()
    }
  }

  handleBodyScroll() {
    this.syncHorizontalScroll(this.bodyScrollTarget.scrollLeft)
  }

  handleIntersection(entries) {
    const entry = entries[0]
    this.element.classList.toggle("is-pinned", entry.boundingClientRect.top < 0)
  }

  syncHorizontalScroll(scrollLeft) {
    this.headerTableTarget.style.transform = `translateX(${-scrollLeft}px)`
  }

  syncHeaderWidth() {
    const firstRow = this.bodyScrollTarget.querySelector("tbody tr")
    if (!firstRow) return

    const widths = Array.from(firstRow.children).map((cell) => cell.getBoundingClientRect().width)
    const frozenWidth = widths[0] || 0
    const scrollableWidth = widths.slice(1).reduce((sum, w) => sum + w, 0)

    this.headerFrozenTarget.style.width = `${frozenWidth}px`
    this.headerScrollTarget.style.width = `${Math.max(0, this.bodyScrollTarget.clientWidth - frozenWidth)}px`
    this.headerTableTarget.style.width = `${scrollableWidth}px`

    const frozenTh = this.headerFrozenTarget.querySelector("th")
    if (frozenTh) {
      frozenTh.style.width = `${frozenWidth}px`
      frozenTh.style.minWidth = `${frozenWidth}px`
    }

    const scrollThs = this.headerTableTarget.querySelectorAll("th")
    scrollThs.forEach((th, index) => {
      const w = widths[index + 1]
      if (w == null) return
      th.style.width = `${w}px`
      th.style.minWidth = `${w}px`
    })

    this.element.style.setProperty("--frozen-column-width", `${frozenWidth}px`)
    this.element.style.setProperty("--scrollable-header-width", `${scrollableWidth}px`)
  }
}
