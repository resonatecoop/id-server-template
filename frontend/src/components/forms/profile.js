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

const SwaggerClient = require('swagger-client')
const CountrySelect = require('../select-country-list')
const RoleSwitcher = require('./roleSwitcher')
const inputField = require('../../elements/input-field')

// help text or link
const helpText = (text, href) => {
  const attrs = {
    class: 'link underline f5 dark-gray tr',
    href: href
  }
  if (href.startsWith('http')) {
    attrs.target = '_blank'
  }
  return html`
    <div class="flex justify-end mt2">
      <a ${attrs}>${text}</a>
    </div>
  `
}

// CheckBox component class
class CheckBox extends Component {
  constructor (id, state, emit) {
    super(id)

    this.emit = emit
    this.state = state

    this.local = state.components[id] = {}

    this.local.checked = 'off'

    this.validator = validateFormdata()
    this.local.form = this.validator.state
  }

  createElement (props = {}) {
    this.local.form = props.form || this.local.form || this.validator.state
    this.onchange = props.onchange // optional callback

    this.local.checked = props.value ? 'on' : 'off'

    const values = this.local.form.values

    values[props.name] = this.local.checked

    const attrs = {
      checked: this.local.checked === 'on' ? 'checked' : false,
      id: props.id || props.name,
      required: false,
      onchange: (e) => {
        this.local.checked = e.target.checked ? 'on' : 'off'
        values[props.name] = this.local.checked
        e.target.setAttribute('checked', e.target.checked ? 'checked' : false)

        typeof this.onchange === 'function' && this.onchange(this.local.checked === 'on')
      },
      value: values[props.name],
      class: 'o-0',
      style: 'width:0;height:0;',
      name: props.name,
      type: 'checkbox'
    }

    if (props.disabled) {
      attrs.disabled = 'disabled'
    }

    return inputField(html`<input ${attrs}>`, this.local.form)({
      prefix: 'flex flex-column mb3',
      disabled: props.disabled,
      labelText: props.labelText || '',
      labelIconName: 'check',
      inputName: props.name,
      helpText: props.helpText,
      displayErrors: true
    })
  }

  update () {
    return false
  }
}

// AccountForm class
class AccountForm extends Component {
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
        const payload = {
          email: this.local.data.email || '',
          displayName: this.local.data.displayName || '',
          membership: this.local.data.member || '',
          newsletter: this.local.data.newsletterNotification ? 'subscribe' : '',
          shares: this.local.shares || '',
          credits: this.local.credits || ''
        }

        response = await fetch('', {
          method: 'PUT',
          headers: {
            Accept: 'application/json',
            'X-CSRF-Token': csrfToken
          },
          body: new URLSearchParams(payload)
        })

        const status = response.status
        const contentType = response.headers.get('content-type')

        if (status >= 400 && contentType && contentType.indexOf('application/json') !== -1) {
          const { error } = await response.json()
          this.local.error.message = error
          this.local.machine.emit('request:error')
        } else {
          this.emit('notify', { message: 'Your account info has been successfully updated' })

          this.local.machine.emit('request:resolve')

          response = await response.json()

          const { data } = response

          if (data.success_redirect_url) {
            setTimeout(() => {
              window.location = data.success_redirect_url
            }, 0)
          }
        }
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

      this.local.machine.emit(`form:${this.local.form.valid ? 'valid' : 'invalid'}`)
    })

    this.validator = validateFormdata()
    this.local.form = this.validator.state
    this.local.shares = 0
  }

  createElement (props = {}) {
    this.local.data = this.local.data || props.data

    const values = this.local.form.values

    for (const [key, value] of Object.entries(this.local.data)) {
      values[key] = value
    }

    return html`
      <div class="flex flex-column flex-auto">
        ${this.state.cache(Form, 'account-form-update').render({
          id: 'account-form',
          method: 'POST',
          action: '',
          buttonText: this.state.profile.complete ? 'Update' : 'Next',
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
              component: this.state.cache(RoleSwitcher, 'role-switcher').render({
                help: true,
                value: this.state.profile.role,
                onChangeCallback: async (value) => {
                  const specUrl = new URL('/user/user.swagger.json', 'https://' + process.env.API_DOMAIN)

                  this.swaggerClient = await new SwaggerClient({
                    url: specUrl.href,
                    authorizations: {
                      bearer: 'Bearer ' + this.state.token
                    }
                  })

                  const roles = [
                    'superadmin',
                    'admin',
                    'tenantadmin',
                    'label', // 4
                    'artist', // 5
                    'user' // 6
                  ]

                  await this.swaggerClient.apis.Users.ResonateUser_UpdateUser({
                    id: this.state.profile.id, // user-api user uuid
                    body: {
                      role_id: roles.indexOf(value) + 1
                    }
                  })
                }
              })
            },
            {
              type: 'text',
              name: 'displayName',
              required: true,
              readonly: this.local.data.displayName ? 'readonly' : false,
              label: 'Display name',
              help: this.local.data.displayName ? helpText('Change your display name', '/profile') : '',
              placeholder: 'Name'
            },
            {
              type: 'email',
              label: 'E-mail',
              help: helpText('Change your email', '/account-settings'),
              readonly: 'readonly', // can't change email address here
              disabled: true
            },
            {
              component: this.state.cache(CountrySelect, 'update-country').render({
                country: this.state.profile.country || '',
                onchange: async (props) => {
                  const { country, code } = props

                  let response = await fetch('')

                  const csrfToken = response.headers.get('X-CSRF-Token')

                  response = await fetch('', {
                    method: 'PUT',
                    headers: {
                      Accept: 'application/json',
                      'X-CSRF-Token': csrfToken
                    },
                    body: new URLSearchParams({
                      country: code
                    })
                  })

                  if (response.status >= 400) {
                    throw new Error('Something went wrong')
                  }

                  this.state.profile.country = country
                }
              })

            },
            {
              component: this.state.cache(CheckBox, 'newsletter-notification').render({
                id: 'newsletterNotification',
                name: 'newsletterNotification',
                value: this.local.data.newsletterNotification,
                form: this.local.form,
                labelText: html`
                  <dl>
                    <dt class="f5">Subscribe to our newsletter</dt>
                    <dd class="f6 ma0">We would like to keep in touch using your email address. Is that OK?</dd>
                  </dl>
                `,
                helpText: helpText('About privacy', 'https://community.resonate.is/docs?search=privacy&topic=1863'),
                onchange: (value) => {
                  this.local.data.newsletterNotification = value
                }
              })
            }
            // {
            //   type: 'text',
            //   name: 'fullName',
            //   required: false,
            //   placeholder: 'Full name'
            // },
            // {
            //   type: 'text',
            //   name: 'firstName',
            //   required: false,
            //   placeholder: 'First name'
            // },
            // {
            //   type: 'text',
            //   name: 'lastName',
            //   required: false,
            //   placeholder: 'Last name'
            // }
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
    this.validator.field('displayName', { required: true }, (data) => {
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

module.exports = AccountForm
