const html = require('choo/html')
const Nanocomponent = require('choo/component')
const assert = require('nanoassert')
const rangeSlider = require('@resonate/rangeslider')

class ProgressBar extends Nanocomponent {
  constructor (id, state, emit) {
    super(id)

    this.state = state
    this.emit = emit
    this.local = state.components[id] = {}

    this.local.progress = 0
    this.local.progressState = 'idle'

    this.createSeeker = this.createSeeker.bind(this)
  }

  createSeeker (el) {
    rangeSlider.create(el, {
      min: 0,
      max: 100,
      value: this.local.progress,
      step: 0.0001,
      rangeClass: 'progressBar',
      disabledClass: 'progressBar--disabled',
      fillClass: 'progressBar__fill',
      bufferClass: 'progressBar__buffer',
      backgroundClass: 'progressBar__background',
      handleClass: 'progressBar__handle'
    })

    return el.rangeSlider
  }

  createElement (props) {
    assert(typeof props.progress, 'number', 'ProgressBar: progress must be a number')

    this.local.progress = props.progress
    this.local.progressState = props.progressState

    if (!this.local.slider) {
      this._element = html`
        <div class="relative">
          <input id="progressBar" disabled="disabled" type="range" />
        </div>
      `
    }

    return this._element
  }

  load (el) {
    el.removeAttribute('unresolved')
    this.local.slider = this.createSeeker(el.querySelector('#progressBar'))
  }

  update (props) {
    if (props.progressState !== this.local.progressState) {
      this.element.classList.add(`progressBar__${this.local.progressState}`)
      this.element.classList.add(`progressBar__${props.progressState}`)
      this.local.progressState = props.progressState
    }
    return props.progress !== this.local.progress
  }
}

module.exports = ProgressBar
