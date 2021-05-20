/* global fetch */

const html = require('choo/html')
const Component = require('choo/component')
const Form = require('./generic')
const logger = require('nanologger')
const log = logger('form:updatePassword')

const isEmpty = require('validator/lib/isEmpty')
const isLength = require('validator/lib/isLength')
const validateFormdata = require('validate-formdata')
const nanostate = require('nanostate')
const PasswordMeter = require('../password-meter')
const zxcvbnAsync = require('zxcvbn-async')

class UpdatePasswordForm extends Component {
  constructor (id, state, emit) {
    super(id)

    this.emit = emit
    this.state = state

    this.local = Object.create({
      machine: nanostate.parallel({
        form: nanostate('idle', {
          idle: { submit: 'submitted' },
          submitted: { valid: 'data', invalid: 'error' },
          data: { reset: 'idle', submit: 'submitted' },
          error: { reset: 'idle', submit: 'submitted', invalid: 'error' }
        }),
        request: nanostate('idle', {
          idle: { start: 'loading' },
          loading: { resolve: 'data', reject: 'error' },
          data: { start: 'loading' },
          error: { start: 'loading', stop: 'idle' }
        }),
        loader: nanostate('off', {
          on: { toggle: 'off' },
          off: { toggle: 'on' }
        })
      })
    })

    this.local.data = {}
    this.local.error = {}

    this.local.machine.on('form:reset', () => {
      this.validator = validateFormdata()
      this.local.form = this.validator.state
    })

    this.local.machine.on('request:start', () => {
      this.loaderTimeout = setTimeout(() => {
        this.local.machine.emit('loader:toggle')
      }, 300)
    })

    this.local.machine.on('request:reject', () => {
      this.emit('notify', { type: 'error', message: 'Something went wrong' })

      clearTimeout(this.loaderTimeout)
    })

    this.local.machine.on('request:resolve', () => {
      this.emit('notify', { type: 'success', message: 'Password changed!' })

      clearTimeout(this.loaderTimeout)
    })

    this.local.machine.on('form:valid', async () => {
      log.info('Form is valid')

      try {
        this.local.machine.emit('request:start')

        let response = await fetch('')

        const csrfToken = response.headers.get('X-CSRF-Token')

        response = await fetch('/password', {
          method: 'PUT',
          headers: {
            Accept: 'application/json',
            'X-CSRF-Token': csrfToken
          },
          body: new URLSearchParams({
            password: this.local.data.password,
            password_new: this.local.data.password_new,
            password_confirm: this.local.data.password_confirm
          })
        })

        this.local.machine.state.loader === 'on' && this.local.machine.emit('loader:toggle')

        const status = response.status
        const contentType = response.headers.get('content-type')

        if (status >= 400 && contentType && contentType.indexOf('application/json') !== -1) {
          const { error } = await response.json()
          this.local.error.message = error
          return this.local.machine.emit('request:error')
        }

        this.local.machine.emit('request:resolve')
      } catch (err) {
        console.log(err)
        this.local.error.message = err.message
        this.local.machine.emit('request:reject')
      }
    })

    this.local.machine.on('form:invalid', () => {
      log.info('Form is invalid')

      const invalidInput = document.querySelector('.invalid')

      if (invalidInput) {
        invalidInput.focus({ preventScroll: false }) // focus to first invalid input
      }
    })

    this.local.machine.on('form:submit', () => {
      log.info('Form has been submitted')

      const form = this.element.querySelector('form')

      for (const field of form.elements) {
        const isRequired = field.required
        const name = field.name || ''
        const value = field.value || ''

        if (isRequired) {
          this.validator.validate(name, value)
        }
      }

      this.rerender()

      if (this.local.form.valid) {
        return this.local.machine.emit('form:valid')
      }

      return this.local.machine.emit('form:invalid')
    })

    this.validator = validateFormdata()
    this.local.form = this.validator.state
  }

  createElement (props) {
    return html`
      <div class="flex flex-column flex-auto pb6">
        ${this.state.cache(Form, 'password-update-form').render({
          id: 'password-update-form',
          method: 'POST',
          action: '',
          buttonText: 'Update my password',
          validate: (props) => {
            this.local.data[props.name] = props.value
            this.validator.validate(props.name, props.value)
            this.rerender()
          },
          form: this.local.form || {
            changed: false,
            valid: true,
            pristine: {},
            required: {},
            values: {},
            errors: {}
          },
          submit: () => {
            this.local.machine.emit('form:submit')
          },
          fields: [
            {
              type: 'password',
              id: 'password',
              autocomplete: 'on',
              name: 'password',
              placeholder: 'Current password'
            },
            {
              type: 'password',
              id: 'password_new',
              autocomplete: 'on',
              name: 'password_new',
              placeholder: 'New password',
              help: (value) => {
                return this.state.cache(PasswordMeter, 'password-meter').render({
                  password: value
                })
              }
            },
            {
              type: 'password',
              id: 'password_confirm',
              autocomplete: 'on',
              name: 'password_confirm',
              placeholder: 'Password confirmation'
            }
          ]
        })}
      </div>
    `
  }

  load () {
    const zxcvbn = zxcvbnAsync.load({
      sync: true,
      libUrl: 'https://cdn.jsdelivr.net/npm/zxcvbn@4.4.2/dist/zxcvbn.js',
      libIntegrity: 'sha256-9CxlH0BQastrZiSQ8zjdR6WVHTMSA5xKuP5QkEhPNRo='
    })

    this.validator.field('password', { required: !!this.local.token }, (data) => {
      if (isEmpty(data)) return new Error('Current password is required')
      if (/[À-ÖØ-öø-ÿ]/.test(data)) return new Error('Current password may contain unsupported characters. You should ask for a password reset.')
    })
    this.validator.field('password_new', (data) => {
      if (isEmpty(data)) return new Error('New password is required')
      if (data === this.local.data.password) return new Error('Current password and new password are identical')
      const { score, feedback } = zxcvbn(data)
      if (score < 3) {
        return new Error(feedback.warning || (feedback.suggestions.length ? feedback.suggestions[0] : 'Password is too weak'))
      }
      if (!isLength(data, { max: 72 })) {
        return new Error('Password length should not be more than 72 characters')
      }
    })
    this.validator.field('password_confirm', (data) => {
      if (isEmpty(data)) return new Error('Password confirmation is required')
      if (data !== this.local.data.password_new) return new Error('Password mismatch')
    })
  }

  update () {
    return false
  }
}

module.exports = UpdatePasswordForm
