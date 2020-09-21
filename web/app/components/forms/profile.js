const html = require('choo/html')
const Component = require('choo/component')
const input = require('@resonate/input-element')
const Button = require('@resonate/button-component')
const messages = require('./messages')
const Dialog = require('@resonate/dialog-component')
const isEqual = require('is-equal-shallow')
const logger = require('nanologger')
const log = logger('form:updateProfile')

const isEmpty = require('validator/lib/isEmpty')
const isEmail = require('validator/lib/isEmail')
const validateFormdata = require('validate-formdata')
const nanostate = require('nanostate')
const inputField = require('../../elements/input-field')

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

        const response = await this.state.api.profile.update({
          email: this.local.data.email,
          nickname: this.local.data.nickname
        })

        if (response.status === 'ok') {
          const dialog = this.state.cache(Dialog, 'profile-updated-dialog')

          const dialogEl = dialog.render({
            title: 'Your profile has been updated.',
            prefix: 'dialog-default dialog--sm pa3',
            onClose: e => {
              dialog.destroy()
            },
            content: html`
              <div class="flex flex-column">
                <p class="lh-copy f5 b">Changes should be taking effect immediatly.</p>

                <div class="flex">
                  <div class="flex items-center">
                  </div>
                  <div class="flex flex-auto w-100 justify-end">
                    <div class="flex items-center">
                      <input class="bg-black white f5 b pv2 ph3 ma0 grow" type="submit" value="Ok">
                    </div>
                  </div>
                </div>
              </div>
            `
          })

          document.body.appendChild(dialogEl)
        }

        if (response.status !== 'ok') {
          this.emit('notify', { type: 'error', message: 'Profile not updated' })
        }

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
  }

  createElement (props = {}) {
    this.local.data = this.local.data || props.data

    const pristine = this.local.form.pristine
    const errors = this.local.form.errors
    const values = this.local.form.values

    for (const [key, value] of Object.entries(this.local.data)) {
      values[key] = value
    }

    const submitButton = new Button('update-profile-button', this.state, this.emit)
    const disabled = (this.local.machine.state.form === 'submitted' && this.local.form.valid) || !this.local.form.changed

    return html`
      <div class="flex flex-column flex-auto pb6">
        ${messages(this.state, this.local.form)}

        <form novalidate onsubmit=${(e) => {
          e.preventDefault()
          this.local.machine.emit('form:submit')
        }}>
          <fieldset id="account_informations" class="ma0 pa0 bn flex flex-column w-100">
            <legend class="f3 mb3">Account information</legend>

            ${inputField(input({
              name: 'email',
              invalid: errors.email && !pristine.email,
              value: values.email,
              onchange: (e) => {
                this.validator.validate(e.target.name, e.target.value)
                this.local.data[e.target.name] = e.target.value
                this.rerender()
              }
            }), this.local.form)({
              prefix: 'mb3',
              labelText: 'Email',
              inputName: 'email',
              displayErrors: true
            })}

            ${inputField(input({
              name: 'displayName',
              value: values.displayName,
              invalid: errors.displayName && !pristine.displayName,
              onchange: (e) => {
                this.validator.validate(e.target.name, e.target.value)
                this.local.data[e.target.name] = e.target.value
                this.rerender()
              }
            }), this.local.form)({
              prefix: 'mb3',
              labelText: 'Profile name',
              inputName: 'displayName',
              displayErrors: true
            })}
          </fieldset>

          ${submitButton.render({
            type: 'submit',
            prefix: `bg-white ba bw b--dark-gray f5 b pv3 ph5 ${!disabled ? 'grow' : ''}`,
            text: 'Update',
            disabled: disabled,
            style: 'none',
            size: 'none'
          })}
        </form>
      </div>
    `
  }

  load () {
    this.validator.field('email', (data) => {
      if (isEmpty(data)) return new Error('Email is required')
      if (!isEmail(data)) return new Error('Email is invalid')
    })
    this.validator.field('displayName', (data) => {
      if (isEmpty(data)) return new Error('Name is required')
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
