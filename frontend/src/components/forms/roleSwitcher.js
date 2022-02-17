const Component = require('choo/component')
const html = require('choo/html')
const icon = require('@resonate/icon-element')
const morph = require('nanomorph')

// RoleSwitcher component class
class RoleSwitcher extends Component {
  /***
   * Create role switcher component
   * @param {String} id - The role switcher component id (unique)
   * @param {Number} state - The choo app state
   * @param {Function} emit - Emit event on choo app
   */
  constructor (id, state, emit) {
    super(id)

    this.emit = emit
    this.state = state

    this.local = state.components[id] = {}

    this.local.value = ''

    this.local.items = [
      { value: 'user', name: 'I\'m a listener' }, // for both artists and listeners
      { value: 'artist', name: 'I\'m an artist' }
      // { value: 'label', name: 'I am a label' }
    ]

    this.handleKeyPress = this.handleKeyPress.bind(this)
    this.updateSelection = this.updateSelection.bind(this)
  }

  /***
   * Create role switcher component element
   * @param {Object} props - The role switcher component props
   * @param {String} props.value - Selected value
   * @returns {HTMLElement}
   */
  createElement (props = {}) {
    this.local.help = props.help
    this.local.value = props.value
    this.onChangeCallback = typeof props.onChangeCallback === 'function'
      ? props.onChangeCallback
      : this.onChangeCallback

    return html`
      <div class="mb3">
        ${this.renderItems()}
        ${this.local.help
          ? html`<p class="f5 lh-copy">Your account type change should take effect at your next log in.</p>`
          : ''
        }
      </div>
    `
  }

  renderItems () {
    return html`
      <div class="items flex flex-auto w-100">
        ${this.local.items.map((item, index) => {
          const { value, name } = item

          const id = 'role-item-' + index

          // item attrs
          const attrs = {
            class: `flex flex-auto w-100 ${index < this.local.items.length - 1 ? ' mr3' : ''}`
          }

          const checked = value === this.local.value

          // input attrs
          const attrs2 = {
            onchange: this.updateSelection,
            id: id,
            tabindex: -1,
            name: 'role',
            type: 'radio',
            disabled: item.hidden ? 'disabled' : false,
            checked: checked,
            value: value
          }

          // label attrs
          const attrs3 = {
            class: `flex items-center justify-center fw4 pv2 w-100 grow bw f5${checked ? ' pr2' : ''}`,
            style: 'outline:solid 1px var(--near-black);outline-offset:-1px',
            tabindex: '0',
            onkeypress: this.handleKeyPress,
            for: id
          }

          return html`
            <div ${attrs}>
              <input ${attrs2}>
              <label ${attrs3}>
                ${checked
                  ? html`
                    <div class="flex flex-shrink-0 justify-center bg-white items-center w2 h2">
                      ${icon('check', { size: 'sm', class: 'fill-transparent' })}
                    </div>
                    `
                  : ''}
                ${name}
              </label>
            </div>
          `
        })}
      </div>
    `
  }

  updateSelection (e) {
    const val = e.target.value
    this.local.value = val
    morph(this.element.querySelector('.items'), this.renderItems())
    this.onChangeCallback(this.local.value)
  }

  handleKeyPress (e) {
    if (e.keyCode === 13) {
      e.preventDefault()
      e.target.control.checked = !e.target.control.checked
      const val = e.target.control.value
      this.local.value = val
      morph(this.element.querySelector('.items'), this.renderItems())
      this.onChangeCallback(this.local.value)
    }
  }

  handleSubmit (e) {
    e.preventDefault()

    this.onChangeCallback(this.local.value)
  }

  onSubmit () {}

  update (props) {
    return props.value !== this.local.value
  }
}

module.exports = RoleSwitcher
