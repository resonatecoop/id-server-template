/* global fetch */

const html = require('choo/html')
const Component = require('choo/component')
const Form = require('./generic')

const isEqual = require('is-equal-shallow')
const logger = require('nanologger')
const log = logger('form:updateProfile')

const isEmpty = require('validator/lib/isEmpty')
const isEmail = require('validator/lib/isEmail')
const validateFormdata = require('validate-formdata')
const nanostate = require('nanostate')

class ProfileForm extends Component {
  constructor (id, state, emit) {
    super(id)

    this.emit = emit
    this.state = state

    this.local = state.components[id] = Object.create({
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
      clearTimeout(this.loaderTimeout)
    })

    this.local.machine.on('request:resolve', () => {
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
            email: this.local.data.email,
            nickname: this.local.data.displayName
          })
        })

        const status = response.status
        const contentType = response.headers.get('content-type')

        if (status >= 400 && contentType && contentType.indexOf('application/json') !== -1) {
          const { error } = await response.json()
          this.local.error.message = error
          return this.local.machine.emit('request:error')
        }

        this.local.machine.emit('request:resolve')
      } catch (err) {
        this.local.machine.emit('request:reject')
        console.log(err)
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

  createElement (props = {}) {
    this.local.data = this.local.data || props.data

    const values = this.local.form.values

    for (const [key, value] of Object.entries(this.local.data)) {
      values[key] = value
    }

    return html`
      <div class="flex flex-column flex-auto">
        ${this.state.cache(Form, 'update-profile-form').render({
          id: 'profile-form',
          method: 'POST',
          action: '',
          buttonText: 'Update my profile',
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
          submit: (data) => {
            this.local.machine.emit('form:submit')
          },
          fields: [
            {
              type: 'email',
              placeholder: 'E-mail'
            },
            {
              type: 'text',
              name: 'displayName',
              placeholder: 'Display name',
              help: html`<p class="ma0 mt1 lh-copy f7">Your artist name, nickname or label name</p>`
            }
          ]
        })}
      </div>
    `
  }

  load () {
    this.validator.field('email', (data) => {
      if (isEmpty(data)) return new Error('Email is required')
      if (!isEmail(data)) return new Error('Email is invalid')
    })
    this.validator.field('displayName', (data) => {
      if (isEmpty(data)) return new Error('Display name is required')
    })
  }

  update (props) {
    if (!isEqual(props.data, this.local.data)) {
      this.local.data = props.data
      return true
    }
    return false
  }
}

module.exports = ProfileForm
