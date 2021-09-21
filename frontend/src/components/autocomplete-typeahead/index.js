const html = require('choo/html')
const Component = require('choo/component')
const validateFormdata = require('validate-formdata')
const input = require('@resonate/input-element')
const icon = require('@resonate/icon-element')
const button = require('@resonate/button')
const Nanobounce = require('nanobounce')
const nanobounce = Nanobounce()

// Autocomplete typeahead input component class
class AutocompleteTypeaheadInput extends Component {
  /***
   * Create an autocomplete typeahead input component
   * @param {String} id - The select country list component id (unique)
   * @param {Number} state - The choo app state
   * @param {Function} emit - Emit event on choo app
   */
  constructor (id, state, emit) {
    super(id)

    this.emit = emit
    this.state = state

    this.local = state.components[id] = {}

    this.addItem = this.addItem.bind(this)
    this.removeItem = this.removeItem.bind(this)

    this.local.results = ['Artist 1', 'Artist 2', 'Artist 4'] // for fuzzy search test
    this.validator = validateFormdata()
    this.local.form = this.validator.state
  }

  /***
   * Create autocomplete typeahead input component element
   * @param {Object} props - The autocomplete typeahead input component props
   * @param {String} props.title - Display title above list
   * @param {Array<String>} props.items - The current data
   * @param {String} props.placeholder - Custom placeholder for input
   * @param {Function} props.eachItem - List item function for map
   * @returns {HTMLElement}
   */
  createElement (props) {
    this.validator = props.validator || this.validator

    this.local.form = props.form || this.local.form || {
      changed: false,
      valid: true,
      pristine: {},
      required: {},
      values: {},
      errors: {}
    }

    this.local.title = props.title || ''
    this.local.items = props.items || []
    this.local.placeholder = props.placeholder || ''

    this.eachItem = typeof props.eachItem === 'function' ? props.eachItem.bind(this) : this.eachItem

    const pristine = this.local.form.pristine
    const errors = this.local.form.errors
    const values = this.local.form.values

    const results = this.local.results.filter((item) => {
      if (!values.q) return false
      const string = item.toLowerCase()
      const val = values.q.toLowerCase()
      return string.includes(val)
    }).map((result, index) => {
      const selected = this.local.items.includes(result)
      const selectedClass = selected ? 'o-50' : 'o-100'

      return html`
        <li onclick=${(e) => this.addItem(result)} class="flex flex-auto items-center striped--near-white relative bb b--black-20 bw1">
          ${button({ disabled: !!selected, classList: selectedClass, iconName: 'add', iconSize: 'sm' })}
          <span class="${selectedClass}">
            ${result}
          </span>
        </li>
      `
    })

    const items = this.local.items.map((item, index) => {
      return html`
        <li class="flex items-center relative pv2">
          ${this.eachItem(item, index)}
          ${button({
            onClick: (e) => this.removeItem(index),
            prefix: 'bg-transparent bn ml2 pa0 grow absolute right-1',
            iconName: 'close-fat',
            iconSize: 'xs',
            style: 'none',
            size: 'none'
          })}
        </li>
      `
    })

    return html`
      <div class="typeahead-component flex flex-column">
        <div class="relative">
          <label class="search-label flex absolute left-1 z-1" for="search">
            ${icon('search', { class: 'fill-white', size: 'sm' })}
            <span class="clip">Search</span>
          </label>
          ${input({
            type: 'search',
            name: 'q',
            value: values.q,
            autocomplete: 'off', // disable native autocomplete
            placeholder: this.local.placeholder,
            onKeyUp: (e) => {
              if (e.key === 'Escape') {
                values.q = ''
                this.rerender()
              }
            },
            onInput: (e) => {
              const value = e.target.value
              values.q = value
              nanobounce(() => {
                this.rerender()
              })
            },
            onchange: (e) => {
              this.validator.validate(e.target.name, e.target.value)
              this.rerender()
            }
          })}
          <div tabindex=0 class="typeahead">
            <ul style="display:${results.length ? 'block' : 'none'}" class="absolute z-1 w-100 bg-white flex flex-column list ma0 pa0 bl bt br b--black-20 bw1">
              ${results}
            </ul>
          </div>
        </div>

        <h5 class="f5 b">${this.local.title}</h5>
        <ul class="list ma0 pa0">
          ${items}
        </ul>

        <p class="ma0 pa0 message warning">
          ${errors.q && !pristine.q ? errors.q.message : ''}
        </p>

      </div>
    `
  }

  // noop
  eachItem () {}

  load () {
    this.validator.field('q', { required: false }, (data) => {
    })
  }

  addItem (value, error) {
    if (value && !this.local.items.includes(value) && !error) {
      this.local.items.push(value)
      this.local.form.values.q = '' // clear
      this.rerender()
    }
  }

  removeItem (index) {
    if (index > -1) {
      this.local.items.splice(index, 1)
      this.rerender()
    }
  }

  update () {
    return true
  }
}

module.exports = AutocompleteTypeaheadInput
