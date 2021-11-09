const Component = require('choo/component')
const compare = require('nanocomponent/compare')
const html = require('choo/html')
const icon = require('@resonate/icon-element')
const morph = require('nanomorph')
const imagePlaceholder = require('../../lib/image-placeholder')

// ProfileSwitcher component class
// [Profile switcher for Artists and Labels only...  multiple tabs, initially one only, scrolling if necessary...  If Label, label tab shown first in different colour, followed by artists on label ]
class ProfileSwitcher extends Component {
  /***
   * Create profile switcher component
   * @param {String} id - The profile switcher component id (unique)
   * @param {Number} state - The choo app state
   * @param {Function} emit - Emit event on choo app
   */
  constructor (id, state, emit) {
    super(id)

    this.emit = emit
    this.state = state

    this.local = state.components[id] = {}

    this.handleKeyPress = this.handleKeyPress.bind(this)
    this.updateSelection = this.updateSelection.bind(this)
  }

  /***
   * Create profile switcher component element
   * @param {Object} props - The profile switcher component props
   * @param {String} props.value - Selected value
   * @returns {HTMLElement}
   */
  createElement (props = {}) {
    this.local.value = props.value
    this.local.ownedGroups = props.ownedGroups || []
    this.onChangeCallback = typeof props.onChangeCallback === 'function'
      ? props.onChangeCallback
      : this.onChangeCallback
    this.local.items = this.local.ownedGroups.map((item) => {
      return {
        value: item.id,
        name: item.displayName,
        banner: item.banner,
        avatar: item.avatar
      }
    })

    console.log('props:', props)

    return html`
      <div class="mb5">
        ${this.renderItems()}
      </div>
    `
  }

  renderItems () {
    return html`
      <div class="items ml-3 mr-3">
        <div class="cf flex flex-wrap">
          ${this.local.items.map((item, index) => {
            const { value, name, avatar } = item

            const id = 'usergroup-item-' + index

            // input attrs
            const attrs = {
              onchange: this.updateSelection,
              id: id,
              tabindex: -1,
              name: 'usergroup',
              type: 'radio',
              disabled: item.hidden ? 'disabled' : false,
              checked: value === this.local.value,
              value: value
            }

            // label attrs
            const attrs2 = {
              class: 'flex flex-column fw4',
              style: 'outline:solid 1px var(--near-black);outline-offset:0px',
              tabindex: '0',
              onkeypress: this.handleKeyPress,
              for: id
            }

            // item background attrs
            const attrs3 = {
              class: 'flex items-end pv3 aspect-ratio--object z-1',
              style: `background: url(${avatar || imagePlaceholder(400, 400)}) center center / cover no-repeat;`
            }

            return html`
              <div class="fl w-33 grow ph3">
                <input ${attrs}>
                <label ${attrs2}>
                  <div class="aspect-ratio aspect-ratio--1x1">
                    <div ${attrs3}>
                      <div class="flex flex-shrink-0 justify-center bg-white ba bw b--mid-gray items-center w2 h2 ml2">
                        ${icon('check', { size: 'sm', class: 'fill-transparent' })}
                      </div>
                      <span class="absolute bottom-0" style="transform:translateY(100%)">${name}</span>
                    </div>
                  </div>
                </label>
              </div>
            `
          })}
        </div>
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
    return compare(this.local.ownedGroups, props.ownedGroups) ||
      this.local.value !== props.value
  }
}

module.exports = ProfileSwitcher
