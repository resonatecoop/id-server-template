/* global fetch */

const html = require('choo/html')
const Component = require('choo/component')
const nanostate = require('nanostate')
const Form = require('./generic')
const isLength = require('validator/lib/isLength')
const isEmpty = require('validator/lib/isEmpty')
const validateFormdata = require('validate-formdata')
const PasswordMeter = require('../password-meter')
const zxcvbnAsync = require('zxcvbn-async')

class PasswordResetUpdatePassword extends Component {
  constructor (id, state, emit) {
    super(id)

    this.emit = emit
    this.state = state

    this.local = state.components[id] = Object.create({
      machine: nanostate.parallel({
        request: nanostate('idle', {
          idle: { start: 'loading' },
          loading: { resolve: 'data', reject: 'error', reset: 'idle' },
          data: { reset: 'idle', start: 'loading' },
          error: { reset: 'idle', start: 'loading' }
        }),
        loader: nanostate('off', {
          on: { toggle: 'off' },
          off: { toggle: 'on' }
        })
      })
    })

    this.local.error = {}
    this.local.success = {}
    this.local.data = {}

    this.local.machine.on('request:error', () => {
      if (this.element) this.rerender()
    })

    this.local.machine.on('request:loading', () => {
      if (this.element) this.rerender()
    })

    this.local.machine.on('loader:toggle', () => {
      if (this.element) this.rerender()
    })

    this.local.machine.transitions.request.event('error', nanostate('error', {
      error: { start: 'loading' }
    }))

    this.local.machine.on('request:noResults', () => {
      if (this.element) this.rerender()
    })

    this.local.machine.transitions.request.event('noResults', nanostate('noResults', {
      noResults: { start: 'loading' }
    }))

    this.validator = validateFormdata()
    this.form = this.validator.state
  }

  createElement (props) {
    this.local.token = props.token

    const message = {
      loading: html`<p class="status w-100 pa1">Loading...</p>`,
      error: html`<p class="status bg-yellow w-100 black pa1">${this.local.error.message}</p>`,
      data: '',
      noResults: html`<p class="status bg-yellow w-100 black pa1">An error occured.</p>`
    }[this.local.machine.state.request]

    return html`
      <div class="flex flex-column flex-auto">
        ${message}
        ${this.state.cache(Form, 'password-reset-update-form').render({
          id: 'password-reset-update-password',
          method: 'POST',
          action: '',
          buttonText: 'Update your password',
          validate: (props) => {
            this.local.data[props.name] = props.value
            this.validator.validate(props.name, props.value)
            this.rerender()
          },
          form: this.form || {
            changed: false,
            valid: true,
            pristine: {},
            required: {},
            values: {},
            errors: {}
          },
          fields: [
            {
              type: 'password',
              id: 'password_new',
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
              name: 'password_confirm',
              placeholder: 'Password confirmation'
            }
          ],
          submit: async (data) => {
            if (this.local.machine.state === 'loading') {
              return
            }

            const loaderTimeout = setTimeout(() => {
              this.local.machine.emit('loader:toggle')
            }, 1000)

            try {
              this.local.machine.emit('request:start')

              let response = await fetch('')

              const csrfToken = response.headers.get('X-CSRF-Token')

              const payload = {
                token: this.local.token,
                password_new: data.password_new.value,
                password_confirm: data.password_confirm.value
              }

              response = await fetch('', {
                method: 'PUT',
                credentials: 'include',
                headers: {
                  Accept: 'application/json',
                  'X-CSRF-Token': csrfToken
                },
                body: new URLSearchParams(payload)
              })

              const isRedirected = response.redirected

              if (isRedirected) {
                window.location.href = response.url
              }

              this.local.machine.state.loader === 'on' && this.local.machine.emit('loader:toggle')

              const status = response.status
              const contentType = response.headers.get('content-type')

              if (status >= 400 && contentType && contentType.indexOf('application/json') !== -1) {
                const { error } = await response.json()
                this.local.error.message = error
                return this.local.machine.emit('request:error')
              }

              const { message } = await response.json()

              if (status === 202) {
                this.emit('notify', { message })
                this.emit(this.state.events.PUSHSTATE, '/login')
              }

              this.local.machine.emit('request:resolve')
            } catch (err) {
              this.local.error.message = err.message
              this.local.machine.emit('request:reject')
              this.emit('error', err)
            } finally {
              clearTimeout(loaderTimeout)
            }
          }
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

    this.validator.field('password_new', (data) => {
      if (isEmpty(data)) {
        return new Error('A strong password is very important')
      }
      if (!isLength(data, { min: 9 })) {
        return new Error('Password length should not be less than 9 characters')
      }
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

module.exports = PasswordResetUpdatePassword
