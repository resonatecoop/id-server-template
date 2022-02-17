const Component = require('choo/component')
const compare = require('nanocomponent/compare')
const html = require('choo/html')
const icon = require('@resonate/icon-element')
const morph = require('nanomorph')
const imagePlaceholder = require('@resonate/svg-image-placeholder')
const NIL_UUID = require('../../lib/nil')

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
    this.local.usergroups = props.usergroups || []
    this.onChangeCallback = typeof props.onChangeCallback === 'function'
      ? props.onChangeCallback
      : this.onChangeCallback
    this.local.items = this.local.usergroups.map((item) => {
      return {
        value: item.id,
        name: item.displayName,
        banner: item.banner,
        avatar: item.avatar
      }
    })

    return html`
      <div class="mb5">
        ${this.renderItems()}
      </div>
    `
  }

  renderItems () {
    const length = this.local.items.length

    const attrs = {
      method: 'POST',
      onsubmit: (e) => {
        e.preventDefault()

        this.onChangeCallback()
      },
      novalidate: 'novalidate',
      action: '/'
    }

    return html`
      <form ${attrs}>
        <div class="items overflow-x-auto overflow-y-hidden bg-light-gray bw bb b--mid-gray">
          <div class="cf flex${length <= 3 ? ' justify-center' : ''}">
            ${this.local.items.map((item, index) => {
              const { value, name, avatar } = item

              const id = 'usergroup-item-' + index
              const checked = value === this.local.value

              // input attrs
              const attrs = {
                onchange: this.updateSelection,
                id: id,
                tabindex: -1,
                name: 'usergroup',
                type: 'radio',
                checked: checked,
                value: value
              }

              // label attrs
              const attrs2 = {
                class: 'flex flex-column fw4',
                style: 'outline:solid 1px var(--near-black);outline-offset:0px',
                tabindex: '0',
                title: 'Select profile',
                onkeypress: this.handleKeyPress,
                for: id
              }

              const src = avatar !== NIL_UUID
                ? `https://${process.env.STATIC_HOSTNAME}/images/${avatar}-x300.jpg`
                : imagePlaceholder(400, 400)

              // item background attrs
              const attrs3 = {
                class: 'flex items-end pb2 aspect-ratio--object z-1',
                style: `background: url(${src}) center center / cover no-repeat;`
              }

              return html`
                <div class="fl flex flex-column justify-center flex-shrink-0 w4 pt2 pb4 ph3">
                  <input ${attrs}>
                  <label ${attrs2}>
                    <div class="aspect-ratio aspect-ratio--1x1">
                      <div ${attrs3}>
                        <div class="flex flex-shrink-0 justify-center items-center ml2">
                          ${icon('circle', { size: 'sm', class: 'fill-transparent' })}
                        </div>
                        <span class="absolute truncate w-100 f5 bottom-0${checked ? ' b' : ''}" style="transform:translateY(100%)">${name}</span>
                      </div>
                    </div>
                  </label>
                </div>
              `
            })}
            <div class="fl flex justify-center items-center flex-shrink-0 w4">
              <button type="submit" title="Create new profile" class="bg-white ba b--mid-gray br-pill w3 h3 mb3 grow">
                <div class="flex items-center justify-center">
                  ${icon('add', { size: 'sm' })}
                </div>
              </button>
            </div>
          </div>
        </div>
      </form>
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
    return compare(this.local.usergroups, props.usergroups) ||
      this.local.value !== props.value
  }
}

module.exports = ProfileSwitcher
