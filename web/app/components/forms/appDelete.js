/* global fetch */

const html = require('choo/html')
const Component = require('choo/component')
const input = require('@resonate/input-element')
const Button = require('@resonate/button-component')
const logger = require('nanologger')
const log = logger('form:deleteApp')

const isEmpty = require('validator/lib/isEmpty')
const validateFormdata = require('validate-formdata')
const nanostate = require('nanostate')
const inputField = require('../../elements/input-field')

class AppDelete extends Component {
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
      this.emit('notify', { type: 'error', message: 'Something went wrong' })

      clearTimeout(this.loaderTimeout)
    })

    this.local.machine.on('request:resolve', () => {
      clearTimeout(this.loaderTimeout)
    })

    this.local.machine.on('form:valid', async () => {
      log.info('Form is valid')

      try {
        this.local.machine.emit('request:start')

        // send request to delete app
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
            _method: 'DELETE',
            client_id: this.local.data.key,
            application_name: this.local.data.application_name
          })
        })

        this.local.machine.emit('request:resolve')
      } catch (err) {
        this.local.machine.emit('request:reject')
        this.emit('error', err)
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
    this.local.data = {}
  }

  createElement (props = {}) {
    const pristine = this.local.form.pristine
    const errors = this.local.form.errors
    const values = this.local.form.values

    this.local.data.key = props.key
    this.local.applicationName = props.applicationName

    const submitButton = new Button('delete-app-button', this.state, this.emit)
    const disabled = (this.local.machine.state.form === 'submitted' && this.local.form.valid) || !this.local.form.changed

    return html`
      <div class="flex flex-column flex-auto pb6">
        <h2 class="lh-title f3 fw1">Delete app</h2>

        <form novalidate onsubmit=${(e) => {
          e.preventDefault()
          this.local.machine.emit('form:submit')
        }}>
          <div class="flex flex-column">
            ${inputField(input({
              name: 'application_name',
              invalid: errors.application_name && !pristine.application_name,
              type: 'text',
              value: values.application_name,
              placeholder: this.local.applicationName,
              onchange: (e) => {
                this.validator.validate(e.target.name, e.target.value)
                this.local.data[e.target.name] = e.target.value
                this.rerender()
              }
            }), this.local.form)({
              prefix: 'mb3',
              labelText: 'Please confirm the application name (security measure)',
              inputName: 'application_name',
              displayErrors: true
            })}

            <label for="delete-client">Permanently delete ${this.local.applicationName} app</label>

            <div class="mb3">
              ${submitButton.render({
                id: 'delete-client',
                type: 'submit',
                prefix: 'bg-red white dib grow ba bw b--near-black b pv2 ph4 flex-shrink-0 f5',
                text: 'Delete app',
                disabled: disabled,
                style: 'none',
                size: 'none'
              })}
            </div>
          </div>
        </form>
      </div>
    `
  }

  load () {
    this.validator.field('application_name', (data) => {
      if (isEmpty(data)) return new Error('Application name is required')
    })
  }

  update (props) {
    return false
  }
}

module.exports = AppDelete
