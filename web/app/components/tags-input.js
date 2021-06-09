const html = require('choo/html')
const Component = require('choo/component')
const validateFormdata = require('validate-formdata')
const input = require('@resonate/input-element')
const isEmpty = require('validator/lib/isEmpty')
const button = require('@resonate/button')
const compare = require('nanocomponent/compare')
const morph = require('nanomorph')

class ItemsInput extends Component {
  constructor (id, state, emit) {
    super(id)

    this.emit = emit
    this.state = state
    this.local = state.components[id] = {}

    this.local.items = []

    this.renderForm = this.renderForm.bind(this)
    this.renderItems = this.renderItems.bind(this)
    this.removeItem = this.removeItem.bind(this)
    this.addItem = this.addItem.bind(this)

    this._update = this._update.bind(this)
    this._onchange = this._onchange.bind(this)

    this.validator = validateFormdata()

    this.local.form = this.validator.state
  }

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

    this.onchange = typeof props.onchange === 'function' ? props.onchange.bind(this) : this._onchange
    this.local.items = props.items || []
    this.local.inputName = props.inputName || 'tags'
    this.local.required = props.required
    this.local.placeholder = props.placeholder || ''

    const { values } = this.local.form

    values[this.local.inputName] = this.local.items.join(',')

    return html`
      <div class="flex flex-column">
        ${this.renderForm(this.local.form)}

        ${this.renderErrors(this.local.form)}

        ${this.renderItems(this.local.items)}
      </div>
    `
  }

  renderForm (form) {
    const { values, errors } = form

    return html`
      <div class="form flex items-center mb1">
        ${input({
          type: 'text',
          name: this.local.inputName,
          placeholder: this.local.placeholder,
          autocomplete: 'off',
          required: this.local.required,
          onKeyPress: (e) => {
            if (e.keyCode === 13) {
              e.preventDefault()
              this.validator.validate(e.target.name, e.target.value)
              this._update()
              this.addItem(values[this.local.inputName], errors[this.local.inputName])
            }
          },
          value: values[this.local.inputName],
          onchange: (e) => {
            this.validator.validate(e.target.name, e.target.value)
            this._update()
          }
        })}
        ${button({
          onclick: (e) => {
            this.addItem(values[this.local.inputName], errors[this.local.inputName])
          },
          prefix: 'db bg-white bw b--black-20 h-100 ml1 pa3 grow',
          style: 'none',
          size: 'none',
          iconName: 'add',
          iconSize: 'sm'
        })}
        ${!this.local.required ? html`<p class="lh-copy f5 pl2 grey">Optional</p>` : ''}
      </div>
    `
  }

  renderErrors (form) {
    const { errors, pristine } = form

    return html`
      <p class="errors ma0 pa0 f5 lh-copy red">
        ${errors[this.local.inputName] && !pristine[this.local.inputName]
          ? errors[this.local.inputName].message
          : ''
        }
      </p>
    `
  }

  renderItems (items) {
    return html`
      <ul class="items flex flex-wrap list ma0 pa0 mt2">
        ${items.map((item, index) => {
          return html`
            <li class="flex items-center bg-black-10 mr2 mb2 ph2 pv2">
              <span class="f5">${item}</span>
              ${button({
                onclick: (e) => this.removeItem(index),
                prefix: 'bg-transparent bn ml2 pa0 grow',
                iconName: 'close-fat',
                iconSize: 'xxs',
                style: 'none',
                size: 'none'
              })}
            </li>
          `
        })}
      </ul>
    `
  }

  _update () {
    morph(this.element.querySelector('.items'), this.renderItems(this.local.items))
    morph(this.element.querySelector('.errors'), this.renderErrors(this.local.form))
    morph(this.element.querySelector('.form'), this.renderForm(this.local.form))
  }

  load () {
    this.validator.field(this.local.inputName, { required: this.local.required }, (data) => {
      if (this.local.required && isEmpty(data)) {
        return new Error(`Adding ${this.local.inputName} is required`)
      }
    })
  }

  _onchange () {}

  removeItem (index) {
    if (index > -1) {
      this.local.items.splice(index, 1)
      this.local.form.values[this.local.inputName] = ''
      this.onchange(this.local.items)
      this._update()
    }
  }

  addItem (value, error) {
    if (value && !this.local.items.includes(value) && !error) {
      this.local.items.push(value)
      this.local.form.values[this.local.inputName] = ''
      this.onchange(this.local.items)
      this._update()
    }
  }

  update (props) {
    return props.form.changed ||
      compare(props.items, this.local.items)
  }
}

module.exports = ItemsInput
