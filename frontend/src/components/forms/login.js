/* global fetch */

const html = require('choo/html')
const Component = require('choo/component')
const nanostate = require('nanostate')
const Form = require('./generic')
const isEmail = require('validator/lib/isEmail')
const isEmpty = require('validator/lib/isEmpty')
const validateFormdata = require('validate-formdata')

class Login extends Component {
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

    this.reset = this.reset.bind(this)

    this.validator = validateFormdata()
    this.form = this.validator.state
  }

  createElement (props) {
    const message = {
      loading: html`<p class="status white w-100 pa2">Loading...</p>`,
      error: html`<p class="status bg-yellow w-100 black pa1">${this.local.error.message}</p>`,
      data: '',
      noResults: html`<p class="status bg-yellow w-100 black pa1">Wrong email or password</p>`
    }[this.local.machine.state.request]

    const confirmEmail = this.state.query.confirm
      ? html`<p class="status bg-yellow w-100 black pa1">Check your mailbox for email confirmation</p>`
      : ''

    const form = this.state.cache(Form, 'login-form').render({
      id: 'login',
      method: 'POST',
      action: '',
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
      buttonText: 'Log In',
      altButton: html`
        <p class="f5 lh-copy">Don't have an account? <a class="link b" href="/join">Join</a>.</p>
      `,
      fields: [
        { type: 'email', autofocus: true, placeholder: 'E-mail' },
        {
          type: 'password',
          placeholder: 'Password',
          help: html`
            <div class="flex justify-end">
              <a href="/password-reset" class="link underline lightGrey f7 ma0 mt1">
                Forgot your password?
              </a>
            </div>
          `
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
              'X-CSRF-Token': csrfToken,
              Pragma: 'no-cache',
              'Cache-Control': 'no-cache'
            },
            body: new URLSearchParams({
              email: data.email.value,
              password: data.password.value
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

          this.local.machine.emit('request:resolve')
        } catch (err) {
          this.local.error.message = err.message
          this.local.machine.emit('request:reject')
          this.emit('error', err)
        } finally {
          clearTimeout(loaderTimeout)
        }
      }
    })

    return html`
      <div class="flex flex-column flex-auto">
        ${confirmEmail}
        ${message}
        ${form}
      </div>
    `
  }

  unload () {
    this.reset()
  }

  reset () {
    this.validator = validateFormdata()
    this.form = this.validator.state
  }

  load () {
    this.validator.field('email', data => {
      if (isEmpty(data)) return new Error('Email is required')
      if (!(isEmail(data))) return new Error('This is not valid email address')
    })

    this.validator.field('password', data => {
      if (isEmpty(data)) return new Error('Password is required')
    })
  }

  update () {
    return false
  }
}

module.exports = Login
