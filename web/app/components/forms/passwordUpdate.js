/* global fetch */

const html = require('choo/html')
const Component = require('choo/component')
const Form = require('./generic')
const Button = require('@resonate/button-component')
const messages = require('./messages')
const Dialog = require('@resonate/dialog-component')
const logger = require('nanologger')
const log = logger('form:updatePassword')

const isEmpty = require('validator/lib/isEmpty')
const isLength = require('validator/lib/isLength')
const validateFormdata = require('validate-formdata')
const nanostate = require('nanostate')

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

        const response = await this.state.api.profile.updatePassword(this.local.data)

        if (response.status === 'ok') {
          const dialog = this.state.cache(Dialog, 'logout-dialog')

          const dialogEl = dialog.render({
            title: 'Your password has been changed.',
            prefix: 'dialog-default dialog--sm pa3',
            onClose: e => {
              if (e.target.returnValue === 'Log out') {
                window.location = `https://${process.env.APP_DOMAIN}/api/user/logout`
              }

              dialog.destroy()
            },
            content: html`
              <div class="flex flex-column">
                <p class="lh-copy f5 b">Do you want to log out now?</p>

                <div class="flex">
                  <div class="flex items-center">
                    <input class="bg-white black ba bw b--near-black f5 b pv2 ph3 ma0 grow" type="submit" value="Later">
                  </div>
                  <div class="flex flex-auto w-100 justify-end">
                    <div class="flex items-center">
                      <input class="bg-black white f5 b pv2 ph3 ma0 grow" type="submit" value="Log out">
                    </div>
                  </div>
                </div>
              </div>
            `
          })

          document.body.appendChild(dialogEl)
        }

        if (response.status !== 'ok') {
          this.emit('notify', { type: 'error', message: response.message })
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

  createElement (props) {
    const submitButton = new Button('update-profile-button', this.state, this.emit)
    const disabled = (this.local.machine.state.form === 'submitted' && this.local.form.valid) || !this.local.form.changed

    return html`
      <div class="flex flex-column flex-auto pb6">
        ${messages(this.state, this.local.form)}

        <form novalidate onsubmit=${(e) => {
          e.preventDefault()
          this.local.machine.emit('form:submit')
        }}>
          ${this.state.cache(Form, 'password-update-form').render({
            id: 'password',
            method: 'POST',
            action: '',
            buttonText: 'Update my password',
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
                name: 'password',
                placeholder: 'Current password'
              },
              {
                type: 'password',
                name: 'password_new',
                placeholder: 'New password'
              },
              {
                type: 'password',
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

                response = await fetch('', {
                  method: 'POST',
                  credentials: 'include',
                  headers: {
                    Accept: 'application/json',
                    'X-CSRF-Token': csrfToken
                  },
                  body: new URLSearchParams({
                    password: data.password.value,
                    password_new: data.password_new.value,
                    password_confirm: data.password_confirm.value
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
    this.validator.field('password', { required: !!this.local.token }, (data) => {
      if (isEmpty(data)) return new Error('Current password is required')
      if (new RegExp(/[À-ÖØ-öø-ÿ]/).test(data)) return new Error('Current password contain unsupported characters. You should ask for a password reset.')
    })
    this.validator.field('password_new', (data) => {
      if (isEmpty(data)) return new Error('New password is required')
      if (!isLength(data, { min: 10 })) return new Error('New password is too short')
      if (data === this.local.data.password) return new Error('Current password and new password are identical')
      if (new RegExp(/[À-ÖØ-öø-ÿ]/).test(data)) return new Error('New password contain unsupported characters (accented chars such as À-ÖØ-öø-ÿ)')
    })
    this.validator.field('password_confirm', (data) => {
      if (isEmpty(data)) return new Error('Password confirmation is required')
      if (data !== this.local.data.password_new) return new Error('Password mismatch')
    })
  }

  unload () {
    if (this.local.machine.state.form !== 'idle') {
      this.local.machine.emit('form:reset')
    }
  }

  update (props) {
    return false
  }
}

module.exports = UpdatePasswordForm
