/* global fetch */

const html = require('choo/html')
const Component = require('choo/component')
const nanostate = require('nanostate')
const isEmpty = require('validator/lib/isEmpty')
const isLength = require('validator/lib/isLength')
const isUUID = require('validator/lib/isUUID')
const validateFormdata = require('validate-formdata')

const input = require('@resonate/input-element')
const textarea = require('../../elements/textarea')
const messages = require('./messages')

const Uploader = require('../image-upload')
const Links = require('../links-input')
const inputField = require('../../elements/input-field')

// BasicInfoForm class
class BasicInfoForm extends Component {
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
        })
      })
    })

    this.local.machine.on('request:start', () => {
    })

    this.local.machine.on('request:reject', () => {
    })

    this.local.machine.on('request:resolve', () => {
    })

    this.local.machine.on('form:valid', async () => {
      try {
        this.local.machine.emit('request:start')

        let response = await fetch('')

        const csrfToken = response.headers.get('X-CSRF-Token')

        response = await fetch('', {
          method: 'PUT',
          headers: {
            Accept: 'application/json',
            'X-CSRF-Token': csrfToken
          },
          body: new URLSearchParams({
            nickname: this.local.data.displayName,
            city: this.local.data.city,
            bio: this.local.data.bio
          })
        })

        this.local.machine.emit('request:resolve')
      } catch (err) {
        this.local.machine.emit('request:reject')
        console.log(err)
      }
    })

    this.local.machine.on('form:invalid', () => {
      const invalidInput = document.querySelector('.invalid')

      if (invalidInput) {
        invalidInput.focus({ preventScroll: false }) // focus to first invalid input
      }
    })

    this.local.machine.on('form:submit', () => {
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

    this.local.subscription = 'off' // newsletter subscription

    this.validator = validateFormdata()
    this.form = this.validator.state

    this.handleSubmit = this.handleSubmit.bind(this)

    // form elements
    this.elements = this.elements.bind(this)
  }

  /***
   * Create basic info form component element
   * @returns {HTMLElement}
   */
  createElement () {
    // form attrs
    const attrs = {
      novalidate: 'novalidate',
      onsubmit: this.handleSubmit
    }

    const submitButton = () => {
      // button attrs
      const attrs = {
        type: 'submit',
        class: 'bg-white near-black dib bn b pv3 ph5 flex-shrink-0 f5 grow',
        style: 'outline:solid 1px var(--near-black);outline-offset:-1px',
        text: 'Continue'
      }
      return html`
        <button ${attrs}>
          Continue
        </button>
      `
    }

    return html`
      <div class="flex flex-column">
        ${messages(this.state, this.form)}
        <form ${attrs}>
          ${Object.entries(this.elements())
            .map(([name, el]) => {
              // possibility to filter by name
              return el(this.validator, this.form)
            })}

          ${submitButton()}
        </form>
      </div>
    `
  }

  /**
   * BasicInfoForm elements
   * @returns {Object} The elements object
   */
  elements () {
    return {
      /**
       * Display name, artist name, nickname for user
       * @param {Object} validator Form data validator
       * @param {Object} form Form data object
       */
      name: (validator, form) => {
        const { values, pristine, errors } = form

        const el = input({
          type: 'text',
          name: 'name',
          placeholder: 'Name',
          invalid: errors.name && !pristine.name,
          value: values.name,
          onchange: (e) => {
            validator.validate(e.target.name, e.target.value)
            this.local.data.name = e.target.value
            this.rerender()
          }
        })

        const labelOpts = {
          labelText: 'Name',
          inputName: 'name'
        }

        return inputField(el, form)(labelOpts)
      },
      /**
       * Description/bio for user
       * @param {Object} validator Form data validator
       * @param {Object} form Form data object
       */
      description: (validator, form) => {
        const { values, pristine, errors } = form

        // TODO user inputField func
        return html`
          <div class="mb5">
            <div class="mb1">
              ${textarea({
                name: 'bio',
                maxlength: 200,
                invalid: errors.bio && !pristine.bio,
                placeholder: 'Short bio',
                required: false,
                text: values.bio,
                onchange: (e) => {
                  validator.validate(e.target.name, e.target.value)
                  this.local.data.bio = e.target.value
                  this.rerender()
                }
              })}
            </div>
            <p class="ma0 pa0 message warning">${errors.bio && !pristine.bio ? errors.bio.message : ''}</p>
            <p class="ma0 pa0 f5 grey">${values.bio ? 200 - values.bio.length : 200} characters remaining</p>
          </div>
        `
      },
      /**
       * Upload user profile image
       * @param {Object} validator Form data validator
       * @param {Object} form Form data object
       */
      profilePicture: (validator, form) => {
        const component = this.state.cache(Uploader, this._name + '-profile-picture')
        const el = component.render({
          name: 'profilePicture',
          form: form,
          validator: validator,
          required: true,
          format: { width: 176, height: 99 },
          accept: 'image/jpeg,image/jpg,image/png',
          ratio: '1600x900px'
        })

        const labelOpts = {
          labelText: 'Profile picture',
          inputName: 'profile-picture'
        }

        return inputField(el, form)(labelOpts)
      },
      /**
       * Upload user header image
       * @param {Object} validator Form data validator
       * @param {Object} form Form data object
       */
      headerImage: (validator, form) => {
        const component = this.state.cache(Uploader, this._name + '-header-image')
        const el = component.render({
          name: 'headerImage',
          required: false,
          form: form,
          validator: validator,
          format: { width: 608, height: 147 },
          accept: 'image/jpeg,image/jpg,image/png',
          ratio: '2480x520px',
          direction: 'column',
          onFileUploaded: async (filename) => {
            console.log(filename)
            this.rerender()
          }
        })

        const labelOpts = {
          labelText: 'Header image',
          inputName: 'header-image'
        }

        return inputField(el, form)(labelOpts)
      },
      /**
       * Location for user (city)
       * @param {Object} validator Form data validator
       * @param {Object} form Form data object
       */
      location: (validator, form) => {
        const { values, pristine, errors } = form

        const el = input({
          type: 'text',
          name: 'location',
          invalid: errors.location && !pristine.location,
          placeholder: 'City',
          required: false,
          value: values.location,
          onchange: (e) => {
            validator.validate(e.target.name, e.target.value)
            this.local.data.city = e.target.value
            this.rerender()
          }
        })

        const labelOpts = {
          labelText: 'Location',
          inputName: 'location'
        }

        return inputField(el, form)(labelOpts)
      },
      /**
       * Links for user
       * @param {Object} validator Form data validator
       * @param {Object} form Form data object
       */
      links: (validator, form) => {
        const { values } = form
        const component = this.state.cache(Links, 'links-input')

        const el = component.render({
          form: form,
          validator: validator,
          value: values.links
        })

        const labelOpts = {
          labelText: 'Links',
          inputName: 'links'
        }

        return inputField(el, form)(labelOpts)
      },
      /**
       * Toggle subscription status for newsletter
       * @param {Object} validator Form data validator
       * @param {Object} form Form data object
       */
      newsletter: (validator, form) => {
        const { values } = form

        const attrs = {
          checked: this.local.subscription === 'on' ? 'checked' : false,
          id: 'subscription',
          onchange: (e) => {
            this.local.subscription = e.target.checked ? 'on' : 'off'
            validator.validate('subscription', this.local.subscription)
            this.rerender()
          },
          value: values.subscription,
          class: 'o-0',
          style: 'width:0;height:0;',
          name: 'subscription',
          type: 'checkbox',
          required: 'required'
        }

        return inputField(html`<input ${attrs}>`, form)({
          prefix: 'flex flex-column mb5',
          labelText: 'Subscribe to newsletter',
          labelIconName: 'check',
          inputName: 'subscription',
          displayErrors: true
        })
      }
    }
  }

  /**
   * Basic info form submit handler
   */
  handleSubmit (e) {
    e.preventDefault()

    this.local.machine.emit('form:submit')
  }

  /**
   * Basic info form submit handler
   * @param {HTMLElement} el THe basic info form element
   */
  load (el) {
    this.validator.field('name', (data) => {
      if (isEmpty(data)) return new Error('Name is required')
    })
    this.validator.field('bio', { required: false }, (data) => {
      if (!isLength(data, { min: 0, max: 200 })) return new Error('Bio should be no more than 200 characters')
    })
    this.validator.field('location', { required: false }, (data) => {})
    this.validator.field('subscription', { required: false }, (data) => {
      if (!isEmpty(data) && ['on', 'off'].indexOf(data) === -1) return new Error('Invalid subscription data')
    })
    this.validator.field('profilePicture', (data) => {
      if (isEmpty(data)) return new Error('Profile picture is required')
      if (!isUUID(data, 4)) return new Error('Profile picture ref is invalid')
    })
    this.validator.field('headerImage', { required: false }, (data) => {
      if (!isEmpty(data) && !isUUID(data, 4)) return new Error('Header image ref is invalid')
    })
  }

  /**
   * Basic info form submit handler
   * @returns {Boolean} Should always returns true
   */
  update () {
    return true
  }
}

module.exports = BasicInfoForm
