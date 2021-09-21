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
// const Links = require('../links-input')
const imagePlaceholder = require('../../lib/image-placeholder')
const inputField = require('../../elements/input-field')

const SwaggerClient = require('swagger-client')
const ProfileTypeForm = require('../../components/forms/profile-type')

// ProfileForm class
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
        machine: nanostate('profileType', {
          profileType: { next: 'basicInfo' },
          basicInfo: { next: 'recap', prev: 'profileType' },
          // customInfo: { next: 'recap', prev: 'basicInfo' }, disabled for now
          recap: { prev: 'basicForm' }
        })
      })
    })

    this.local.machine.on('machine:next', () => {
      if (this.element) {
        this.rerender()
      }
    })

    this.local.machine.on('machine:prev', () => {
      if (this.element) {
        this.rerender()
      }
    })

    this.local.machine.on('form:valid', async () => {
      try {
        this.local.machine.emit('request:start')

        const specUrl = new URL('/user/user.swagger.json', 'https://' + process.env.API_DOMAIN)
        const client = await new SwaggerClient({
          url: specUrl.href,
          authorizations: {
            bearer: 'Bearer ' + this.state.token
          }
        })

        if (!this.local.persona.id) {
          const response = await client.apis.Usergroups.ResonateUser_AddUserGroup({
            id: this.state.profile.id,
            body: {
              displayName: this.local.data.displayName,
              description: this.local.data.description,
              shortBio: this.local.data.shortBio,
              address: this.local.data.location,
              avatar: this.local.data.avatar, // uuid
              banner: this.local.data.banner, // uuid
              groupType: 'persona'
            }
          })

          this.local.persona.id = response.body.id
        } else {
          const response = await client.apis.Usergroups.ResonateUser_UpdateUserGroup({
            id: this.local.persona.id, // should be usergroup id
            body: {
              displayName: this.local.data.displayName,
              description: this.local.data.description,
              address: this.local.data.address,
              shortBio: this.local.data.shortBio,
              avatar: this.local.data.avatar,
              banner: this.local.data.banner
            }
          })

          console.log(response.body)
        }

        this.local.machine.emit('machine:next')
      } catch (err) {
        this.local.machine.emit('request:reject')
        console.log(err)
      }
    })

    this.local.machine.on('form:invalid', () => {
      const invalidInput = this.element.querySelector('.invalid')

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

      this.local.machine.emit(`form:${this.form.valid ? 'valid' : 'invalid'}`)
    })

    this.local.data = {}
    this.local.persona = {}

    this.validator = validateFormdata()
    this.form = this.validator.state

    this.handleSubmit = this.handleSubmit.bind(this)
    this.renderForm = this.renderForm.bind(this)

    // form elements
    this.elements = this.elements.bind(this)
  }

  /***
   * Create basic info form component element
   * @returns {HTMLElement}
   */
  createElement (props = {}) {
    // initial persona
    if (!this.local.persona.id) {
      const profile = props.profile
      const persona = profile.ownedGroups.find(ownedGroup => {
        return ownedGroup.groupType === 'persona'
      }) || {
        displayName: profile.nickname,
        description: profile.description || '',
        avatar: profile.avatar['profile_photo-m'] || profile.avatar['profile_photo-l'] || imagePlaceholder(400, 400)
      }

      if (persona.id) {
        this.local.usergroup = persona.id // for updates
        this.local.data.banner = persona.banner
        this.local.data.avatar = persona.avatar
        this.local.data.address = persona.address
        this.local.data.shortBio = persona.shortBio
      }

      this.local.data.description = persona.description
      this.local.data.displayName = persona.displayName

      this.local.persona = persona
    }

    const steps = {
      profileType: () => {
        return this.state.cache(ProfileTypeForm, 'profile-type').render({
          onSubmit: (usergroup) => {
            this.local.usergroupType = usergroup
            this.local.machine.emit('machine:next')
          }
        })
      },
      basicInfo: this.renderForm, // basic infos for everyone
      // customInfo: this.renderCustomInfoForm, // label, artist infos (disabled for now
      recap: renderRecap // recap
    }[this.local.machine.state.machine]

    return html`
      <div class="flex flex-column">
        ${steps()}
      </div>
    `

    function renderRecap () {
      return html`
        <p class="lh-copy fw1 f4">Thank you for completing your profile!</p>
      `
    }
  }

  renderForm () {
    // find first available persona or fallback to available legacy profile
    const persona = this.local.persona
    const values = this.form.values

    for (const [key, value] of Object.entries(this.local.data)) {
      values[key] = value
    }

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
        <div class="mb5">
          <h4 class="lh-title mb2 f4 fw1">Profile</h4>

          <p class="lh-copy f5 ma0 mb2">You are currently editing ${persona.displayName}’s profile.</p>

          <div class="flex items-center pv2">
            <div class="fl w-100 mw3">
              <div class="db aspect-ratio aspect-ratio--1x1 bg-dark-gray bg-dark-gray--dark">
                <div class="aspect-ratio--object cover" style="background:url(${persona.avatar || this.state.profile.avatar['profile_photo-m'] || this.state.profile.avatar['profile_photo-l'] || imagePlaceholder(400, 400)}) center;"></div>
              </div>
            </div>
            <div>
              <span class="pa3">${persona.displayName}</span>
            </div>
          </div>
        </div>
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
      displayName: (validator, form) => {
        const { values, pristine, errors } = form

        const el = input({
          type: 'text',
          name: 'displayName',
          invalid: errors.displayName && !pristine.displayName,
          value: values.displayName,
          onchange: (e) => {
            validator.validate(e.target.name, e.target.value)
            this.local.data.displayName = e.target.value
            this.rerender()
          }
        })

        const labelOpts = {
          labelText: 'Name',
          inputName: 'displayName',
          helpText: 'Your artist name, nickname or label name.',
          displayErrors: true
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

        return html`
          <div class="mb5">
            <div class="mb1">
              ${textarea({
                name: 'description',
                maxlength: 2000,
                invalid: errors.description && !pristine.description,
                placeholder: 'Bio',
                required: false,
                text: values.description,
                onchange: (e) => {
                  validator.validate(e.target.name, e.target.value)
                  this.local.data.description = e.target.value
                  this.rerender()
                }
              })}
            </div>
            <p class="ma0 pa0 message warning">${errors.description && !pristine.description ? errors.description.message : ''}</p>
            <p class="ma0 pa0 f5 dark-gray">${values.description ? 2000 - values.description.length : 2000} characters remaining</p>
          </div>
        `
      },
      /**
       * Short bio
       * @param {Object} validator Form data validator
       * @param {Object} form Form data object
       */
      shortBio: (validator, form) => {
        const { values, pristine, errors } = form

        return html`
          <div class="mb5">
            <div class="mb1">
              ${textarea({
                name: 'shortBio',
                maxlength: 100,
                invalid: errors.shortBio && !pristine.shortBio,
                placeholder: 'Short bio',
                required: false,
                text: values.shortBio,
                onchange: (e) => {
                  validator.validate(e.target.name, e.target.value)
                  this.local.data.shortBio = e.target.value
                  this.rerender()
                }
              })}
            </div>
            <p class="ma0 pa0 message warning">${errors.shortBio && !pristine.shortBio ? errors.shortBio.message : ''}</p>
            <p class="ma0 pa0 f5 dark-gray">${values.shortBio ? 100 - values.shortBio.length : 100} characters remaining</p>
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
          config: 'avatar',
          required: false,
          validator: validator,
          format: { width: 176, height: 99 },
          src: this.local.persona.avatar,
          accept: 'image/jpeg,image/jpg,image/png',
          ratio: '1600x900px',
          archive: this.state.profile.avatar['profile_photo-m'] || this.state.profile.avatar['profile_photo-l'], // last uploaded files, old wp cover photo...
          onFileUploaded: async (filename) => {
            this.local.data.avatar = filename

            if (!this.local.usergroup.id) return

            try {
              const specUrl = new URL('/user/user.swagger.json', 'https://' + process.env.API_DOMAIN)
              const client = await new SwaggerClient({
                url: specUrl.href,
                authorizations: {
                  bearer: 'Bearer ' + this.state.token
                }
              })

              await client.apis.Usergroups.ResonateUser_UpdateUserGroup({
                id: this.local.usergroup.id, // should be usergroup id
                body: {
                  avatar: this.local.data.avatar
                }
              })

              this.emit('notify', { message: 'Profile picture updated', type: 'success' })
            } catch (err) {
              console.log(err)
            }
          }
        })

        const labelOpts = {
          labelText: 'Profile picture',
          labelPrefix: 'f4 fw1 db mb2',
          inputName: 'profile-picture',
          displayErrors: true
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
          form: form,
          config: 'banner',
          required: false,
          validator: validator,
          src: this.local.persona.banner,
          format: { width: 608, height: 147 },
          accept: 'image/jpeg,image/jpg,image/png',
          ratio: '2480x520px',
          direction: 'column',
          archive: this.state.profile.avatar['cover_photo-m'], // last uploaded files, old wp cover photo...
          onFileUploaded: async (filename) => {
            this.local.data.banner = filename

            if (!this.local.usergroup.id) return

            try {
              const specUrl = new URL('/user/user.swagger.json', 'https://' + process.env.API_DOMAIN)
              const client = await new SwaggerClient({
                url: specUrl.href,
                authorizations: {
                  bearer: 'Bearer ' + this.state.token
                }
              })

              await client.apis.Usergroups.ResonateUser_UpdateUserGroup({
                id: this.local.usergroup.id, // should be usergroup id
                body: {
                  banner: this.local.data.banner
                }
              })

              this.emit('notify', { message: 'Profile picture updated', type: 'success' })
            } catch (err) {
              console.log(err)
            }
          }
        })

        const labelOpts = {
          labelText: 'Header image',
          labelPrefix: 'f4 fw1 db mb2',
          inputName: 'header-image',
          displayErrors: true
        }

        return inputField(el, form)(labelOpts)
      }
      /**
       * Address for user (could be a place, city, anywhere)
       * @param {Object} validator Form data validator
       * @param {Object} form Form data object
       */
      /*
      address: (validator, form) => {
        const { values, pristine, errors } = form

        const el = input({
          type: 'text',
          name: 'address',
          invalid: errors.address && !pristine.address,
          placeholder: 'City',
          required: false,
          value: values.address,
          onchange: (e) => {
            validator.validate(e.target.name, e.target.value)
            this.local.data.address = e.target.value
            this.rerender()
          }
        })

        const labelOpts = {
          labelText: 'Location',
          inputName: 'location'
        }

        return inputField(el, form)(labelOpts)
      }
      */
      /**
       * Links for user
       * @param {Object} validator Form data validator
       * @param {Object} form Form data object
       */
      /*
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
      }
      */
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
    this.validator.field('displayName', (data) => {
      if (isEmpty(data)) return new Error('Display name is required')
      if (!isLength(data, { min: 1, max: 100 })) return new Error('Name should be no more than 100 characters')
    })
    this.validator.field('description', { required: false }, (data) => {
      if (!isLength(data, { min: 0, max: 2000 })) return new Error('Description should be no more than 2000 characters')
    })
    this.validator.field('shortBio', { required: false }, (data) => {
      if (!isLength(data, { min: 0, max: 100 })) return new Error('Short bio should be no more than 100 characters')
    })
    /*
    this.validator.field('address', { required: false }, (data) => {
      if (!isLength(data, { min: 0, max: 100 })) return new Error('Location should be no more than 100 characters')
    })
    */
    this.validator.field('profilePicture', { required: false }, (data) => {
      if (!isEmpty(data) && !isUUID(data, 4)) return new Error('Profile picture ref is invalid')
    })
    this.validator.field('headerImage', { required: false }, (data) => {
      if (!isEmpty(data) && !isUUID(data, 4)) return new Error('Header image ref is invalid')
    })
  }

  /**
   * Basic info form submit handler
   * @returns {Boolean} Should always returns true
   */
  update (props) {
    return true
  }
}

module.exports = ProfileForm
