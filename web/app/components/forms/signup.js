/* global fetch */

const html = require('choo/html')
const Component = require('choo/component')
const nanostate = require('nanostate')
const Form = require('./generic')
const isEmail = require('validator/lib/isEmail')
const isEmpty = require('validator/lib/isEmpty')
const isLength = require('validator/lib/isLength')
const validateFormdata = require('validate-formdata')

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
    this.form = this.validator.state
  }

  createElement (props) {
    const message = {
      loading: html`<p class="status white w-100 pa2">Loading...</p>`,
      error: html`<p class="status bg-yellow w-100 black pa1">${this.local.error.message}</p>`,
      data: '',
      noResults: html`<p class="status bg-yellow w-100 black pa1">Wrong email or password</p>`
    }[this.local.machine.state.request]

    return html`
      <div class="flex flex-column flex-auto">
        ${message}
        ${this.state.cache(Form, 'signup-form').render({
          id: 'signup',
          method: 'POST',
          action: '',
          buttonText: 'Sign up',
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
              label: 'E-mail'
            },
            {
              type: 'text',
              name: 'login',
              label: 'Username'
            },
            {
              type: 'password',
              label: 'Password'
            },
            {
              type: 'text',
              name: 'display_name',
              label: 'Name',
              help: html`<p class="ma0 mt1 lh-copy f7">Your artist name, nickname or label name</p>`
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

              if (status === 201) {
                this.emit(this.state.events.PUSHSTATE, '/login')
              }

              this.machine.emit('request:resolve')
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
      if (isEmpty(data)) return new Error('Email is required')
      if (!(isEmail(data))) return new Error('Email is not valid')
    })
    this.validator.field('password', (data) => {
      if (isEmpty(data)) return new Error('Password is required')
      if (!isLength(data, { min: 9, max: 72 })) return new Error('Password should contain between 9 and 72 characters')
    })
    this.validator.field('display_name', (data) => {
      if (isEmpty(data)) return new Error('Name is required')
      if (!isLength(data, { max: 50 })) return new Error('Full name is too long')
    })
    this.validator.field('login', (data) => {
      if (isEmpty(data)) return new Error('Username is required')
      if (!isLength(data, { max: 60 })) return new Error('Username name is too long')
    })
  }

  update () {
    return false
  }
}

module.exports = Signup
