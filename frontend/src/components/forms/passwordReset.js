/* global fetch */

const html = require('choo/html')
const Component = require('choo/component')
const nanostate = require('nanostate')
const Form = require('./generic')
const isEmail = require('validator/lib/isEmail')
const isEmpty = require('validator/lib/isEmpty')
const validateFormdata = require('validate-formdata')

class PasswordReset extends Component {
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
    const message = {
      loading: html`<p class="status w-100 pa1">Loading...</p>`,
      error: html`<p class="status bg-yellow w-100 black pa1">${this.local.error.message}</p>`,
      data: html`<p class="w-100 pa1">${this.local.success.message}</p>`,
      noResults: html`<p class="status bg-yellow w-100 black pa1">An error occured.</p>`
    }[this.local.machine.state.request]

    return html`
      <div class="flex flex-column flex-auto">
        ${message}
        ${this.state.cache(Form, 'password-reset-form').render({
          id: 'password-reset',
          method: 'POST',
          action: '',
          buttonText: 'Reset my password',
          validate: (props) => {
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
              type: 'email',
              label: 'To reset your password, please enter your email address below',
              placeholder: 'Enter your email address'
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

              const payload = { email: data.email.value }

              response = await fetch('', {
                method: 'POST',
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
                this.local.machine.emit('request:error')
              } else {
                const { message } = await response.json()

                this.local.success.message = message

                this.local.machine.emit('request:resolve')

                this.rerender()
              }
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
    this.validator.field('email', (data) => {
      if (isEmpty(data)) return new Error('Please enter your email address')
      if (!(isEmail(data))) return new Error('This is not a valid email address')
    })
  }

  update () {
    return false
  }
}

module.exports = PasswordReset
