import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
  static targets = ["textInput", "wpmSlider", "wpmDisplay", "startBtn", "resetBtn", "wordDisplay", "progress", "timeEstimate", "status"]

  connect() {
    this.words = []
    this.currentIndex = 0
    this.isPlaying = false
    this.intervalId = null

    const savedWpm = localStorage.getItem('speed-reader-wpm')
    if (savedWpm) {
      this.wpmSliderTarget.value = savedWpm
      this.wpmDisplayTarget.textContent = savedWpm
    }

    this.displayWord('')
    this.updateProgress()
    this.updateTimeEstimate()
  }

  disconnect() {
    this.pause()
  }

  // Calculate the optimal recognition point (ORP) - approximately middle of word
  getORP(word) {
    const len = word.length
    if (len <= 1) return 0
    if (len <= 3) return 1
    return Math.floor((len - 1) / 2)
  }

  // Display a word with the focus letter highlighted
  displayWord(word) {
    const display = this.wordDisplayTarget

    while (display.firstChild) {
      display.removeChild(display.firstChild)
    }

    const beforeSpan = document.createElement('span')
    beforeSpan.className = 'before'

    const focusSpan = document.createElement('span')
    focusSpan.className = 'focus'

    const afterSpan = document.createElement('span')
    afterSpan.className = 'after'

    if (word) {
      const orpIndex = this.getORP(word)
      const before = word.slice(0, orpIndex)
      const focus = word[orpIndex]
      const after = word.slice(orpIndex + 1)

      const maxPadding = 10
      const beforePadding = maxPadding - before.length
      const afterPadding = maxPadding - after.length
      const paddedBefore = '\u00A0'.repeat(Math.max(0, beforePadding)) + before
      const paddedAfter = after + '\u00A0'.repeat(Math.max(0, afterPadding))

      beforeSpan.textContent = paddedBefore
      focusSpan.textContent = focus
      afterSpan.textContent = paddedAfter
    }

    display.appendChild(beforeSpan)
    display.appendChild(focusSpan)
    display.appendChild(afterSpan)
  }

  updateProgress() {
    this.progressTarget.textContent = `${this.currentIndex} / ${this.words.length} words`
  }

  updateTimeEstimate() {
    const text = this.textInputTarget.value.trim()
    const wordCount = text ? text.split(/\s+/).filter(w => w.length > 0).length : 0

    if (wordCount === 0) {
      this.timeEstimateTarget.textContent = ''
      return
    }

    const wpm = parseInt(this.wpmSliderTarget.value, 10)
    const totalSeconds = Math.ceil((wordCount / wpm) * 60)
    const minutes = Math.floor(totalSeconds / 60)
    const seconds = totalSeconds % 60

    if (minutes > 0) {
      this.timeEstimateTarget.textContent = `${wordCount} words \u2022 ${minutes}m ${seconds}s at ${wpm} WPM`
    } else {
      this.timeEstimateTarget.textContent = `${wordCount} words \u2022 ${seconds}s at ${wpm} WPM`
    }
  }

  updateStatus(text, className = '') {
    this.statusTarget.textContent = text
    this.statusTarget.className = 'status ' + className
  }

  getInterval() {
    const wpm = parseInt(this.wpmSliderTarget.value, 10)
    return Math.round(60000 / wpm)
  }

  showNextWord() {
    if (this.currentIndex >= this.words.length) {
      this.pause()
      this.updateStatus('Finished', '')
      this.startBtnTarget.textContent = 'Restart'
      return
    }

    this.displayWord(this.words[this.currentIndex])
    this.currentIndex++
    this.updateProgress()
  }

  play() {
    if (this.words.length === 0) return

    this.isPlaying = true
    this.updateStatus('Playing', 'playing')
    this.startBtnTarget.textContent = 'Pause'

    this.showNextWord()
    this.intervalId = setInterval(() => this.showNextWord(), this.getInterval())
  }

  pause() {
    this.isPlaying = false
    this.updateStatus('Paused', 'paused')
    this.startBtnTarget.textContent = 'Resume'

    if (this.intervalId) {
      clearInterval(this.intervalId)
      this.intervalId = null
    }
  }

  initializeReader() {
    const text = this.textInputTarget.value.trim()
    if (!text) {
      alert('Please enter some text first.')
      return
    }

    this.words = text.split(/\s+/).filter(w => w.length > 0)
    this.currentIndex = 0

    this.updateProgress()
    this.displayWord('')
    this.updateStatus('Ready', '')
    this.startBtnTarget.textContent = 'Start'
  }

  // Actions
  start() {
    if (this.words.length === 0) {
      this.initializeReader()
      if (this.words.length > 0) {
        this.play()
      }
    } else if (this.currentIndex >= this.words.length) {
      this.currentIndex = 0
      this.play()
    } else if (this.isPlaying) {
      this.pause()
    } else {
      this.play()
    }
  }

  reset() {
    this.pause()
    this.words = []
    this.currentIndex = 0
    this.displayWord('')
    this.updateProgress()
    this.updateStatus('Ready', '')
    this.startBtnTarget.textContent = 'Start'
  }

  updateWpm() {
    this.wpmDisplayTarget.textContent = this.wpmSliderTarget.value
    localStorage.setItem('speed-reader-wpm', this.wpmSliderTarget.value)
    this.updateTimeEstimate()

    if (this.isPlaying) {
      clearInterval(this.intervalId)
      this.intervalId = setInterval(() => this.showNextWord(), this.getInterval())
    }
  }

  textChanged() {
    this.updateTimeEstimate()
  }

  keydown(event) {
    if (document.activeElement === this.textInputTarget) return

    switch (event.code) {
      case 'Space':
        event.preventDefault()
        this.start()
        break
      case 'ArrowLeft':
        event.preventDefault()
        this.previousWord()
        break
      case 'ArrowRight':
        event.preventDefault()
        this.nextWord()
        break
      case 'ArrowUp':
        event.preventDefault()
        this.wpmSliderTarget.value = Math.min(800, parseInt(this.wpmSliderTarget.value, 10) + 5)
        this.updateWpm()
        break
      case 'ArrowDown':
        event.preventDefault()
        this.wpmSliderTarget.value = Math.max(100, parseInt(this.wpmSliderTarget.value, 10) - 5)
        this.updateWpm()
        break
    }
  }

  previousWord() {
    if (this.words.length === 0) return

    if (this.isPlaying) {
      this.pause()
    }

    this.currentIndex = Math.max(0, this.currentIndex - 2)
    this.displayWord(this.words[this.currentIndex])
    this.currentIndex++
    this.updateProgress()
  }

  nextWord() {
    if (this.words.length === 0) return

    if (this.isPlaying) {
      this.pause()
    }

    if (this.currentIndex < this.words.length) {
      this.displayWord(this.words[this.currentIndex])
      this.currentIndex++
      this.updateProgress()
    }
  }
}
