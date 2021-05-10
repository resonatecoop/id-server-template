/* global fetch */

const html = require('choo/html')
const Component = require('choo/component')
const nanostate = require('nanostate')
const Form = require('./generic')
const isURL = require('validator/lib/isURL')
const isEmpty = require('validator/lib/isEmpty')
const isFQDN = require('validator/lib/isFQDN')
const validateFormdata = require('validate-formdata')

class App extends Component {
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
        ${this.state.cache(Form, 'app-form').render({
          id: 'app-new',
          method: 'POST',
          action: '',
          buttonText: 'Save',
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
              type: 'text',
              name: 'application_name',
              placeholder: 'Enter app name'
            },
            {
              type: 'text',
              name: 'redirect_uri',
              placeholder: 'Redirect URI'
            },
            {
              type: 'text',
              name: 'application_url',
              placeholder: 'Application URL'
            },
            {
              type: 'text',
              name: 'application_hostname',
              placeholder: 'Application Hostname'
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
                application_name: data.application_name.value,
                redirect_uri: data.redirect_uri.value,
                application_url: data.application_url.value,
                application_hostname: data.application_hostname.value
              }

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
                return this.local.machine.emit('request:error')
              }

              const { message } = await response.json()

              this.local.success.message = message

              this.local.machine.emit('request:resolve')

              this.rerender()
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
    this.validator.field('application_name', (data) => {
      if (isEmpty(data)) return new Error('Please enter your app name')
    })
    this.validator.field('redirect_uri', (data) => {
      if (isEmpty(data)) return new Error('Redirect URI is required')
      if (!(isURL(data, { require_protocol: true }))) return new Error('This is not a valid url (protocol is required)')
    })
    this.validator.field('application_url', (data) => {
      if (isEmpty(data)) return new Error('App url is required')
      if (!(isURL(data, { require_protocol: true }))) return new Error('This is not a valid url (protocol is required)')
    })
    this.validator.field('application_hostname', (data) => {
      if (isEmpty(data)) return new Error('A hostname is required')
      if (!(isFQDN(data))) return new Error('This is not valid hostname (fqdn)')
    })
  }

  update () {
    return false
  }
}

module.exports = App
