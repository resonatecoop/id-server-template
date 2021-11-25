/* global fetch */

const html = require('choo/html')
const Component = require('choo/component')
const nanostate = require('nanostate')
const Form = require('./generic')
const isEmail = require('validator/lib/isEmail')
const isEmpty = require('validator/lib/isEmpty')
const isLength = require('validator/lib/isLength')
const validateFormdata = require('validate-formdata')
const PasswordMeter = require('../password-meter')
const CountrySelect = require('../select-country-list')
const zxcvbnAsync = require('zxcvbn-async')
const RoleSwitcher = require('./roleSwitcher')

class Signup extends Component {
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
    this.local.form = this.validator.state
    this.local.role = 'user'
    this.local.roleId = 6
  }

  createElement (props) {
    const message = {
      loading: html`<p class="status w-100 pa1">Loading...</p>`,
      error: html`<p class="status bg-yellow w-100 black pa1">${this.local.error.message}</p>`,
      data: '',
      noResults: html`<p class="status bg-yellow w-100 black pa1">An error occured.</p>`
    }[this.local.machine.state.request]

    return html`
      <div class="flex flex-column flex-auto">
        ${message}
        ${this.state.cache(Form, 'signup-form').render({
          id: 'signup',
          method: 'POST',
          action: '',
          buttonText: 'Sign up',
          altButton: html`
            <p class="f5 lh-copy">Already have an account? <a class="link b" href="/login">Log In</a>.</p>
          `,
          validate: (props) => {
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
          fields: [
            {
              component: this.state.cache(RoleSwitcher, 'role-switcher').render({
                value: this.local.role,
                onChangeCallback: async (value) => {
                  this.local.role = value
                }
              })
            },
            {
              type: 'email',
              placeholder: 'E-mail'
            },
            {
              type: 'password',
              placeholder: 'Password',
              help: (value) => {
                return this.state.cache(PasswordMeter, 'password-meter').render({
                  password: value
                })
              }
            },
            {
              component: this.state.cache(CountrySelect, 'join-country-select').render({
                validator: this.validator,
                form: this.local.form || {
                  changed: false,
                  valid: true,
                  pristine: {},
                  required: {},
                  values: {},
                  errors: {}
                },
                required: true,
                onchange: (e) => {
                  // something changed
                }
              })
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

              response = await fetch('', {
                method: 'POST',
                credentials: 'include',
                headers: {
                  Accept: 'application/json',
                  'X-CSRF-Token': csrfToken
                },
                body: new URLSearchParams({
                  email: data.email.value,
                  password: data.password.value,
                  country: data.country.value,
                  role: this.local.role
                })
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

              if (status === 201) {
                const redirectURL = new URL('/login', 'http://localhost')

                redirectURL.search = new URLSearchParams({
                  confirm: true,
                  login_redirect_uri: '/web/account'
                })

                this.emit(this.state.events.PUSHSTATE, redirectURL.pathname + redirectURL.search)
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

    this.validator.field('email', (data) => {
      if (isEmpty(data)) {
        return new Error('Please tell us your email address')
      }
      if (!isEmail(data)) {
        return new Error('This is not a valid email address')
      }
    })
    this.validator.field('password', (data) => {
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
  }

  update () {
    return false
  }
}

module.exports = Signup
