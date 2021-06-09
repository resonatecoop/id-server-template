const html = require('choo/html')
const Component = require('choo/component')
const isCreditCard = require('validator/lib/isCreditCard')
const isEmpty = require('validator/lib/isEmpty')
const isLength = require('validator/lib/isLength')
const validateFormdata = require('validate-formdata')
const input = require('@resonate/input-element')

class PaymentMethods extends Component {
  constructor (name, state, emit) {
    super(name)

    this.emit = emit
    this.state = state
    this.validator = validateFormdata()
    this.form = this.validator.state
  }

  createElement (props) {
    this.validator = props.validator || this.validator
    this.form = props.form || this.form || {
      changed: false,
      valid: true,
      pristine: {},
      required: {},
      values: {},
      errors: {}
    }

    const pristine = this.form.pristine
    const errors = this.form.errors
    const values = this.form.values

    return html`
      <div>
        <div class="flex flex-column">
          <div class="mb1">
            ${input({
              type: 'text',
              name: 'name',
              invalid: errors.name && !pristine.name,
              autocomplete: 'on',
              placeholder: 'Name on card',
              value: values.name,
              onchange: (e) => {
                this.validator.validate(e.target.name, e.target.value)
                this.rerender()
              }
            })}
          </div>
          <p class="ma0 pa0 message warning">${errors.name && !pristine.name ? errors.name.message : ''}</p>
          <div class="mb1">
            ${input({
              type: 'text',
              name: 'number',
              invalid: errors.number && !pristine.number,
              autocomplete: 'on',
              placeholder: 'Card number',
              min: 13,
              max: 19,
              value: values.number,
              onchange: (e) => {
                this.validator.validate(e.target.name, e.target.value)
                this.rerender()
              }
            })}
          </div>
          <p class="ma0 pa0 message warning">${errors.number && !pristine.number ? errors.number.message : ''}</p>

          <div class="flex">
            <div class="mr1">
              <label for="expiration">Expiration date</label>
              <div class="mb3" style="width:123px">
                ${input({
                  type: 'text',
                  name: 'expiration',
                  autocomplete: 'on',
                  classList: 'tc',
                  invalid: errors.expiration && !pristine.expiration,
                  max: 5,
                  onInput: (e) => {
                    if (e.target.value.length === 2) {
                      e.target.value += '/'
                    }
                  },
                  placeholder: 'MM/YY',
                  value: values.expiration,
                  onchange: (e) => {
                    this.validator.validate(e.target.name, e.target.value)
                    this.rerender()
                  }
                })}
              </div>
              <p class="ma0 pa0 message warning">${errors.expiration && !pristine.expiration ? errors.expiration.message : ''}</p>
            </div>
            <div>
              <label for="cvc">CVC</label>
              <div class="mb1" style="width:123px">
                ${input({
                  type: 'text',
                  name: 'cvc',
                  autocomplete: 'on',
                  invalid: errors.cvc && !pristine.cvc,
                  max: 3,
                  value: values.cvc,
                  onchange: (e) => {
                    this.validator.validate(e.target.name, e.target.value)
                    this.rerender()
                  }
                })}
              </div>
              <p class="ma0 pa0 message warning">${errors.cvc && !pristine.cvc ? errors.cvc.message : ''}</p>
            </div>
          </div>
        </div>
      </div>
    `
  }

  load () {
    this.validator.field('name', (data) => {
      if (isEmpty(data)) return new Error('Card name is required')
    })

    this.validator.field('number', (data) => {
      if (isEmpty(data)) return new Error('Card number is missing')
      if (!isLength(data, { min: 13, max: 19 })) return new Error('Card number length should be between 13 and 19 digits')
      if (!isCreditCard(data)) return new Error('Card number is invalid')
    })

    this.validator.field('expiration', (data) => {
      if (isEmpty(data)) return new Error('Expiration date is required')
      const month = parseInt(data.split('/')[0], 10)
      const year = parseInt(data.split('/')[1], 10)
      if (isNaN(month) || isNaN(year)) return new Error('Invalid date')
      const now = new Date()
      const then = new Date()
      then.setFullYear(Math.floor(new Date().getFullYear() / 100) * 100 + year, month, 1)
      if (then < now) return new Error('Card expiration date has passed')
    })

    this.validator.field('cvc', (data) => {
      if (isEmpty(data)) return new Error('CVC is required')
      if (isNaN(parseInt(data, 10))) return new Error('CVC should be 3 digits')
      if (!isLength(data, { min: 3, max: 3 })) return new Error('CVC should be 3 digits')
    })
  }

  update () {
    return true
  }
}

module.exports = PaymentMethods
