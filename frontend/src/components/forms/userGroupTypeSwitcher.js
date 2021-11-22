const Component = require('choo/component')
const html = require('choo/html')
const icon = require('@resonate/icon-element')
const morph = require('nanomorph')

// UserGroupTypeSwitcher component class
class UserGroupTypeSwitcher extends Component {
  /***
   * Create usergroup type component
   * @param {String} id - The usergroup type component id (unique)
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
      { value: 'persona', name: 'Persona' }, // for both artists and listeners
      { value: 'band', name: 'Band' },
      { value: 'label', name: 'Label' }
    ]

    this.handleKeyPress = this.handleKeyPress.bind(this)
    this.updateSelection = this.updateSelection.bind(this)
  }

  /***
   * Create usergroup type component element
   * @param {Object} props - The usergroup type component props
   * @param {String} props.value - Selected value
   * @returns {HTMLElement}
   */
  createElement (props = {}) {
    this.local.value = props.value
    this.onChangeCallback = typeof props.onChangeCallback === 'function'
      ? props.onChangeCallback
      : this.onChangeCallback

    return html`
      <div class="mb5">
        ${this.renderItems()}
        <p class="f5 lh-copy">Dev only: Changes to to the currently selected usergroup take effect immediatly.</p>
      </div>
    `
  }

  renderItems () {
    return html`
      <div class="items flex flex-auto w-100">
        ${this.local.items.map((item, index) => {
          const { value, name } = item

          const id = 'usergroup-type-item-' + index

          // item attrs
          const attrs = {
            class: `flex flex-auto w-100 ${index < this.local.items.length - 1 ? ' mr3' : ''}`
          }

          // input attrs
          const attrs2 = {
            onchange: this.updateSelection,
            id: id,
            tabindex: -1,
            name: 'usergroup-type',
            type: 'radio',
            disabled: item.hidden ? 'disabled' : false,
            checked: value === this.local.value,
            value: value
          }

          // label attrs
          const attrs3 = {
            class: 'flex items-center fw4 pv3 w-100 grow bw',
            style: 'outline:solid 1px var(--near-black);outline-offset:-1px',
            tabindex: '0',
            onkeypress: this.handleKeyPress,
            for: id
          }

          return html`
            <div ${attrs}>
              <input ${attrs2}>
              <label ${attrs3}>
                <div class="flex flex-shrink-0 justify-center bg-white items-center w2 h2 ml2">
                  ${icon('check', { size: 'sm', class: 'fill-transparent' })}
                </div>
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

module.exports = UserGroupTypeSwitcher
