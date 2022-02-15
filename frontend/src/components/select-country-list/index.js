const html = require('choo/html')
const Component = require('choo/component')
const isEmpty = require('validator/lib/isEmpty')
const validateFormdata = require('validate-formdata')
const icon = require('@resonate/icon-element')

// Select country list component class
class SelectCountryList extends Component {
  /***
   * Create a select country list component
   * @param {String} id - The select country list component id (unique)
   * @param {Number} state - The choo app state
   * @param {Function} emit - Emit event on choo app
   */
  constructor (id, state, emit) {
    super(id)

    this.emit = emit
    this.state = state

    this.local = state.components[id] = {}

    this._onchange = this._onchange.bind(this)
    this.validator = validateFormdata()
    this.form = this.validator.state

    this.nameMap = {}
    this.codeMap = {}
  }

  /***
   * Create select country list component element
   * @param {Object} props - The select country list component props
   * @param {String} props.country - Initial country name or country Alpha-2 code
   * @param {String} props.name - Select element name attribute
   * @param {Object} props.form - Form
   * @param {Object} props.validator - Validator
   * @param {Boolean} props.required - Select element required attribute
   * @param {Function} props.onchange - Optional onchange callback
   * @returns {HTMLElement}
   */
  createElement (props = {}) {
    this.state.countries.forEach(this.mapCodeAndName.bind(this))

    this.local.options = this.state.countries.map(({ code, name }) => {
      return {
        value: code,
        label: name
      }
    }).sort((a, b) => a.label.localeCompare(b.label, 'en', {}))

    this.validator = props.validator || this.validator
    this.form = props.form || this.validator.state

    this.local.required = props.required || false
    this.local.name = props.name || 'country'

    this.onchange = props.onchange // optional callback

    const pristine = this.form.pristine
    const errors = this.form.errors
    const values = this.form.values

    if (!this.local.country) {
      this.local.country = props.country || this.getName(values[this.local.name] || '') || ''
    }

    // select attributes
    const attrs = {
      id: 'select-country',
      required: this.local.required,
      class: 'w-100 bn br0 bg-black white bg-white--dark black--dark bg-black--light white--light pa3',
      onchange: this._onchange,
      name: this.local.name
    }

    return html`
      <div class="mb3">
        <div class="flex flex-auto flex-column">

          ${errors[attrs.name] && !pristine[attrs.name]
            ? html`
              <div class="absolute left-0 ph1 flex items-center" style="top:50%;transform: translate(-100%, -50%);">
                ${icon('info', { class: 'fill-red', size: 'sm' })}
              </div>
            `
            : ''
          }
          <label for="select-country" class="f5 db mb1 dark-gray">
            Country
          </label>
          <select ${attrs}>
            <option value="" selected=${!values[attrs.name]} disabled>Select a country</option>
            ${this.local.options.map(({ value, label, disabled = false }) => {
              const selected = this.local.country === value || this.getCode(this.local.country) === value
              return html`
                <option value=${value} disabled=${disabled} selected=${selected}>
                  ${label}
                </option>
              `
            })}
          </select>
          ${errors[attrs.name] && !pristine[attrs.name]
            ? html`<span class="message f5 pb2">${errors[attrs.name].message}</span>`
            : ''
          }
        </div>
      </div>
    `
  }

  mapCodeAndName (country) {
    this.nameMap[country.name.toLowerCase()] = country.code
    this.codeMap[country.code.toLowerCase()] = country.name
  }

  getCode (name) {
    return this.nameMap[name.toLowerCase()]
  }

  getName (code) {
    return this.codeMap[code.toLowerCase()]
  }

  /**
   * Select element onchange event handler
   * @param {Object} e Event
   */
  async _onchange (e) {
    const value = e.target.value

    this.local.code = value
    this.local.country = this.getName(value)

    this.validator.validate('country', value)
    this.rerender()

    // optional callback
    try {
      typeof this.onchange === 'function' && await this.onchange({
        country: this.local.country,
        code: this.local.code
      })

      this.emit('notify', {
        message: `Location changed to ${this.local.country}`,
        type: 'success'
      })
    } catch (err) {
      this.emit('notify', {
        message: 'Location not changed',
        type: 'error'
      })
    }
  }

  /***
   * Select country list component on load event handler
   * @param {HTMLElement} el - The select country list component element
   */
  load (el) {
    this.validator.field(this.local.name, { required: this.local.required }, (data) => {
      if (this.local.required && isEmpty(data)) {
        return new Error('Please select a country')
      }
    })
  }

  /***
   * Select country list component on update event handler
   * @param {Object} props - Select country list component props
   * @param {String} props.country - The current selected country
   * @param {Object} props.validator - Validator
   * @returns {Boolean} Should update
   */
  update (props) {
    return props.country !== this.local.country ||
      (props.validator && props.validator.changed)
  }
}

module.exports = SelectCountryList
