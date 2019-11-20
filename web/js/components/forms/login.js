const html = require('choo/html')
const Component = require('choo/component')
const nanostate = require('nanostate')
const Form = require('./generic')
const isEmail = require('validator/lib/isEmail')
const isEmpty = require('validator/lib/isEmpty')
const validateFormdata = require('validate-formdata')

/* global fetch */

class Login extends Component {
  constructor (name, state, emit) {
    super(name)

    this.emit = emit
    this.state = state

    this.machine = nanostate('idle', {
      idle: { start: 'loading' },
      loading: { resolve: 'idle', reject: 'error' },
      error: { start: 'loading' }
    })

    this.machine.on('loading', () => {
      this.rerender()
    })

    this.machine.on('error', () => {
      this.rerender()
    })

    this.reset = this.reset.bind(this)

    this.validator = validateFormdata()
    this.form = this.validator.state
  }

  createElement (props) {
    const message = {
      loading: html`<p class="status bg-gray bg--mid-gray--dark black w-100 pa2">Loading...</p>`,
      error: html`<p class="status bg-yellow w-100 black pa1">Wrong email or password</p>`
    }[this.machine.state]

    const form = this.state.cache(Form, 'login-form').render({
      id: 'login',
      method: 'POST',
      action: '', // keep query strings
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
      buttonText: 'Login',
      fields: [
        { type: 'email', autofocus: true, placeholder: 'Email' },
        { type: 'password', placeholder: 'Password', help: html`<div class="flex justify-end"><a href="https://resonate.is/password-reset/" class="lightGrey f7 ma0 pt1 pr2" target="_blank" rel="noopener noreferer">Forgot your password?</a></div>` }
      ],
      submit: async (data) => {
        try {
          const email = data.email.value
          const password = data.password.value

          const response = await fetch('', {
            method: 'POST',
            headers: { 'Content-type': 'application/x-www-form-urlencoded' },
            body: `email=${email}&password=${password}`
          })

          if (response.redirected) {
            window.location.href = response.url
          }
        } catch (err) {
          console.log(err)
        }
      }
    })

    return html`
      <div class="flex flex-column flex-auto">
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
      if (!(isEmail(data))) return new Error('Email is not valid')
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
