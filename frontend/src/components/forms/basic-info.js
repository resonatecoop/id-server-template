const html = require('choo/html')
const Component = require('choo/component')
const nanostate = require('nanostate')
const isEmpty = require('validator/lib/isEmpty')
const isLength = require('validator/lib/isLength')
const isUUID = require('validator/lib/isUUID')
const validateFormdata = require('validate-formdata')
const icon = require('@resonate/icon-element')
const morph = require('nanomorph')

const input = require('@resonate/input-element')
const textarea = require('../../elements/textarea')
const messages = require('./messages')

const button = require('@resonate/button')
const Dialog = require('@resonate/dialog-component')
const Uploader = require('../image-upload')
const AutocompleteTypeahead = require('../autocomplete-typeahead')
// const Links = require('../links-input')
// const Tags = require('../tags-input')

const imagePlaceholder = require('../../lib/image-placeholder')
const inputField = require('../../elements/input-field')

const UserGroupTypeSwitcher = require('../../components/forms/userGroupTypeSwitcher')
const ProfileSwitcher = require('../../components/forms/profileSwitcher')

const SwaggerClient = require('swagger-client')

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
        machine: nanostate('basicInfo', {
          basicInfo: { next: 'customInfo', end: 'recap' },
          customInfo: { next: 'recap', prev: 'basicInfo' },
          recap: { prev: 'customInfo', start: 'basicInfo' }
        })
      })
    })

    this.local.machine.on('machine:next', () => {
      if (!this.element) return
      this.rerenderForm()
      window.scrollTo(0, 0)
    })

    this.local.machine.on('machine:end', () => {
      if (!this.element) return
      this.rerenderForm()
      window.scrollTo(0, 0)
    })

    this.local.machine.on('machine:next', () => {
      if (!this.element) return
      this.rerenderForm()
      window.scrollTo(0, 0)
    })

    this.local.machine.on('machine:prev', () => {
      if (!this.element) return
      this.rerenderForm()
      window.scrollTo(0, 0)
    })

    this.local.machine.on('form:valid', async () => {
      try {
        this.local.machine.emit('request:start')

        await this.getClient(this.state.token)

        if (!this.local.usergroup.id) {
          const response = await this.swaggerClient.apis.Usergroups.ResonateUser_AddUserGroup({
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

          this.local.usergroup = response.body
        } else {
          await this.swaggerClient.apis.Usergroups.ResonateUser_UpdateUserGroup({
            id: this.local.usergroup.id, // should be usergroup id
            body: {
              displayName: this.local.data.displayName,
              description: this.local.data.description,
              address: this.local.data.address,
              shortBio: this.local.data.shortBio
            }
          })
        }

        this.local.machine.emit('request:resolve')
      } catch (err) {
        this.local.machine.emit('request:reject')
        console.log(err)
      } finally {
        this.local.machine.emit('form:reset')
      }
    })

    this.local.machine.on('form:invalid', () => {
      console.log('form is invalid')

      const invalidInput = this.element.querySelector('.invalid')

      if (invalidInput) {
        invalidInput.focus({ preventScroll: false }) // focus to first invalid input
      }
    })

    this.local.machine.on('form:submit', () => {
      console.log('form has been submitted')

      const form = this.element.querySelector('form')

      for (const field of form.elements) {
        const isRequired = field.required
        const name = field.name || ''
        const value = field.value || ''

        if (isRequired) {
          this.validator.validate(name, value)
        }
      }

      this.rerenderForm()

      this.local.machine.emit(`form:${this.local.form.valid ? 'valid' : 'invalid'}`)
    })

    this.local.data = {}
    this.local.usergroup = {}
    this.local.profile = {}

    this.validator = validateFormdata()
    this.local.form = this.validator.state

    this.renderBasicInfoForm = this.renderBasicInfoForm.bind(this)
    this.renderCustomInfoForm = this.renderCustomInfoForm.bind(this)
    this.renderRecap = this.renderRecap.bind(this)
    this.rerenderForm = this.rerenderForm.bind(this)

    // cached swagger client
    this.swaggerClient = null

    this.getClient = this.getClient.bind(this)
    this.setUsergroup = this.setUsergroup.bind(this)
  }

  /**
   * Get swagger client
   */
  async getClient (token) {
    if (this.swaggerClient !== null) {
      return this.swaggerClient
    }

    const specUrl = new URL('/user/user.swagger.json', 'https://' + process.env.API_DOMAIN)

    this.swaggerClient = await new SwaggerClient({
      url: specUrl.href,
      authorizations: {
        bearer: 'Bearer ' + token
      }
    })

    return this.swaggerClient
  }

  /***
   * Create basic info form component element
   * @returns {HTMLElement}
   */
  createElement (props = {}) {
    this.local.profile = props.profile || {}
    this.local.role = props.role

    // initial persona
    if (!this.local.usergroup.id) {
      this.setUsergroup()
    }

    const machine = {
      basicInfo: this.renderBasicInfoForm, // basic infos for everyone
      customInfo: this.renderCustomInfoForm, // label, artist infos
      recap: this.renderRecap // recap
    }[this.local.machine.state.machine]

    return html`
      <div class="flex flex-column">
        ${machine()}
      </div>
    `
  }

  /**
   * Set current usergroup
   */
  setUsergroup (usergroupID) {
    const profile = Object.assign({}, this.local.profile)

    const usergroup = profile.ownedGroups.find(ownedGroup => {
      // try find usergroup by id first
      if (usergroupID) return ownedGroup.id === usergroupID
      // fallback to persona or label groupType
      return ownedGroup.groupType === 'persona' || ownedGroup.groupType === 'label'
    }) || {
      // fallback to older profile data for returning members
      displayName: profile.nickname,
      description: profile.description || '',
      avatar: profile.avatar['profile_photo-m'] || profile.avatar['profile_photo-l'] || imagePlaceholder(400, 400)
    }

    if (usergroup.id) {
      this.local.groupType = usergroup.groupType
      this.local.data.banner = usergroup.banner
      this.local.data.avatar = usergroup.avatar
      this.local.data.address = usergroup.address
      this.local.data.shortBio = usergroup.shortBio
    }

    this.local.data.description = usergroup.description
    this.local.data.displayName = usergroup.displayName

    this.local.usergroup = usergroup
  }

  /**
   * Rerender only base form element
   */
  rerenderForm () {
    const machine = {
      basicInfo: this.renderBasicInfoForm, // basic infos for everyone
      customInfo: this.renderCustomInfoForm, // label, artist infos
      recap: this.renderRecap // recap
    }[this.local.machine.state.machine]

    morph(this.element.querySelector('.base-form'), machine())
  }

  /**
   * Basic info form
   */
  renderBasicInfoForm () {
    // form elements
    const elements = {
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
          onchange: async (e) => {
            validator.validate(e.target.name, e.target.value)
            this.local.data.displayName = e.target.value
            this.rerenderForm()

            if (!this.local.usergroup.id) return

            try {
              await this.getClient(this.state.token)

              await this.swaggerClient.apis.Usergroups.ResonateUser_UpdateUserGroup({
                id: this.local.usergroup.id, // should be usergroup id
                body: {
                  displayName: this.local.data.displayName
                }
              })

              this.emit('notify', { message: 'Display name saved' })
            } catch (err) {
              console.log(err)
              this.emit('notify', { message: 'Failed saving display name' })
            }
          }
        })

        const helpText = this.local.role && this.local.role !== 'user' ? `Your ${this.local.role} name` : 'Your username'

        const labelOpts = {
          labelText: 'Name',
          inputName: 'displayName',
          helpText: helpText,
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
                onchange: async (e) => {
                  validator.validate(e.target.name, e.target.value)
                  this.local.data.description = e.target.value
                  this.rerenderForm()

                  if (!this.local.usergroup.id) return

                  try {
                    await this.getClient(this.state.token)

                    await this.swaggerClient.apis.Usergroups.ResonateUser_UpdateUserGroup({
                      id: this.local.usergroup.id, // should be usergroup id
                      body: {
                        description: this.local.data.description
                      }
                    })

                    this.emit('notify', { message: 'Description saved' })
                  } catch (err) {
                    console.log(err)
                    this.emit('notify', { message: 'Failed saving description' })
                  }
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
                onchange: async (e) => {
                  validator.validate(e.target.name, e.target.value)
                  this.local.data.shortBio = e.target.value
                  this.rerenderForm()

                  if (!this.local.usergroup.id) return

                  try {
                    await this.getClient(this.state.token)

                    await this.swaggerClient.apis.Usergroups.ResonateUser_UpdateUserGroup({
                      id: this.local.usergroup.id, // should be usergroup id
                      body: {
                        shortBio: this.local.data.shortBio
                      }
                    })
                    this.emit('notify', { message: 'Short bio saved' })
                  } catch (err) {
                    console.log(err)
                    this.emit('notify', { message: 'Failed saving short bio' })
                  }
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
          format: { width: 300, height: 300 }, // minimum accepted format values
          src: this.local.usergroup.avatar,
          accept: 'image/jpeg,image/jpg,image/png',
          ratio: '1600x1600px',
          archive: this.state.profile.avatar['profile_photo-m'] || this.state.profile.avatar['profile_photo-l'], // last uploaded files, old wp cover photo...
          onFileUploaded: async (filename) => {
            this.local.data.avatar = filename

            if (!this.local.usergroup.id) return

            try {
              await this.getClient(this.state.token)

              await this.swaggerClient.apis.Usergroups.ResonateUser_UpdateUserGroup({
                id: this.local.usergroup.id, // should be usergroup id
                body: {
                  avatar: this.local.data.avatar
                }
              })

              this.emit('notify', { message: 'Profile picture updated', type: 'success' })
            } catch (err) {
              console.log(err)
              this.emit('notify', { message: 'Profile picture failed to update', type: 'success' })
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
          src: this.local.usergroup.banner,
          format: { width: 608, height: 147 },
          accept: 'image/jpeg,image/jpg,image/png',
          ratio: '2480x520px',
          direction: 'column',
          archive: this.state.profile.avatar['cover_photo-m'], // last uploaded files, old wp cover photo...
          onFileUploaded: async (filename) => {
            this.local.data.banner = filename

            if (!this.local.usergroup.id) return

            try {
              await this.getClient(this.state.token)

              await this.swaggerClient.apis.Usergroups.ResonateUser_UpdateUserGroup({
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
      }//,
      /**
       * Address for user (could be a place, city, anywhere, should enable this later, not supported by user-api yet)
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
          onchange: async (e) => {
            validator.validate(e.target.name, e.target.value)
            this.local.data.address = e.target.value
            this.rerender()

            if (!this.local.usergroup.id) return

            try {
              await this.getClient(this.state.token)

              await this.swaggerClient.apis.Usergroups.ResonateUser_UpdateUserGroup({
                id: this.local.usergroup.id, // should be usergroup id
                body: {
                  address: this.local.data.address
                }
              })

              this.emit('notify', { message: 'Location updated', type: 'success' })
            } catch (err) {
              console.log(err)
            }
          }
        })

        const labelOpts = {
          labelText: 'Location',
          inputName: 'location'
        }

        return inputField(el, form)(labelOpts)
      },
      */
      /**
       * Links for usergroup (enable this later, not supported by user-api yet)
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
      },
      /**
       * Tags for usergroup (enable this later, not supported by user-api yet)
       * @param {Object} validator Form data validator
       * @param {Object} form Form data object
       */
      /*
      tags: (validator, form) => {
        const { values } = form
        const component = this.state.cache(Tags, 'tags-input')

        const el = component.render({
          form: form,
          validator: validator,
          value: values.tags,
          items: ['test']
        })

        const labelOpts = {
          labelText: 'Links',
          inputName: 'links'
        }

        return inputField(el, form)(labelOpts)
      }
      */
    }

    const action = this.local.usergroup.id // usergroup id
      ? 'Update'
      : 'Create'

    return this.renderForm(`${action} your ${this.local.role ? `${this.local.role} ` : ''}profile`, elements)
  }

  /*
   * renderCustomInfoForm
   */
  renderCustomInfoForm () {
    const elements = {
      /**
       * Links for user (artist|label)
       * @param {Object} validator Form data validator
       * @param {Object} form Form data object
       */
      artists: (validator, form) => {
        const component = this.state.cache(AutocompleteTypeahead, 'artists-list')

        const el = component.render({
          form: this.local.form,
          validator: this.validator,
          title: 'Artists',
          eachItem: function (item, index) {
            return html`
              <div onclick=${(e) => {
                e.preventDefault()

                const validator = validateFormdata()
                const form = validator.state

                validator.field('displayName', (data) => {
                  if (isEmpty(data)) return new Error('Display name is required')
                })

                validator.field('role', (data) => {
                  if (isEmpty(data)) return new Error('Role is required')
                })

                const pristine = form.pristine
                const errors = form.errors
                const values = form.values

                const dialogEl = this.state.cache(Dialog, 'member-role').render({
                  title: 'Set member display name and role',
                  onClose: function (e) {
                    e.preventDefault()

                    for (const field of e.target.elements) {
                      const isRequired = field.required
                      const name = field.name || ''
                      const value = field.value || ''

                      if (isRequired) {
                        validator.validate(name, value)
                      }
                    }

                    morph(this.element.querySelector('.content'), this.content())

                    if (this.local.form.valid) {
                      this.close()
                    }
                  },
                  content: html`
                    <div class="content flex flex-column">
                      <p class="ph1">${item}</p>

                      <div class="mb2">
                        <label for="displayName" class="f6">Display Name</label>
                        ${input({
                          type: 'text',
                          name: 'displayName',
                          invalid: errors.displayName && !pristine.displayName,
                          required: 'required',
                          value: values.displayName,
                          onchange: (e) => {
                            validator.validate(e.target.name, e.target.value)
                            // morph(this.element.querySelector('.content'), content())
                          }
                        })}
                        <p class="ma0 pa0 message warning">${errors.displayName && !pristine.displayName ? errors.displayName.message : ''}</p>
                      </div>

                      <div class="mb2">
                        <label for="role" class="f6">Role</label>
                        ${input({
                          type: 'text',
                          name: 'role',
                          required: 'required',
                          placeholder: 'E.g.Bass Guitar',
                          invalid: errors.role && !pristine.role,
                          value: values.role,
                          onchange: (e) => {
                            validator.validate(e.target.name, e.target.value)
                            // morph(this.element.querySelector('.content'), content())
                          }
                        })}
                        <p class="ma0 pa0 message warning">${errors.role && !pristine.role ? errors.role.message : ''}</p>
                      </div>

                      <div class="flex">
                        ${button({ type: 'submit', text: 'Continue' })}
                      </div>
                    </div>
                  `
                })

                document.body.appendChild(dialogEl)
              }}>
              ${item}
            </div>`
          },
          placeholder: 'Members name',
          items: ['AGF']
        })

        const labelOpts = {
          labelText: 'Artists',
          inputName: 'artists'
        }

        return inputField(el, form)(labelOpts)
      }
    }

    const title = `Your label: ${this.local.data.displayName}`

    return this.renderForm(title, elements)
  }

  /*
   * All done with setting up account profile
   */
  renderRecap () {
    return html`
      <div>
        <p class="lh-copy fw1 f4">Thank you for completing your profile!</p>
      </div>
    `
  }

  renderProfileSwitcher () {
    if (!this.local.role || this.local.role === 'user' || this.local.machine.state.machine !== 'basicInfo') return

    return this.state.cache(ProfileSwitcher, 'profile-switcher').render({
      value: this.local.usergroup.id, // currently selected usergroup/persona
      ownedGroups: this.local.profile.ownedGroups,
      onChangeCallback: (usergroupId) => {
        this.setUsergroup(usergroupId)

        this.rerenderForm()
      }
    })
  }

  /**
   * Dev only for role switching
   */
  renderRoleSwitcher () {
    if (process.env.NODE_ENV !== 'development') return

    // groupe type assign
    const groupType = {
      user: 'persona', // listener
      artist: 'persona',
      label: 'label'
    }[this.local.role]

    return this.state.cache(UserGroupTypeSwitcher, 'usergroup-type-switcher').render({
      value: this.local.usergroup.groupType || groupType,
      onChangeCallback: async (groupType) => {
        if (!this.local.usergroup.id) return

        try {
          await this.getClient(this.state.token)

          await this.swaggerClient.apis.Usergroups.ResonateUser_UpdateUserGroup({
            id: this.local.usergroup.id, // should be usergroup id
            body: {
              groupType: groupType
            }
          })

          this.emit('notify', { message: `Usergroup type changed to: ${groupType}` })
        } catch (err) {
          console.log(err)
          this.emit('notify', { message: 'Failed setting group type' })
        }
      }
    })
  }

  /*
   * Render form
   */
  renderForm (title, elements) {
    // find first available persona or fallback to available legacy profile
    const values = this.local.form.values

    for (const [key, value] of Object.entries(this.local.data)) {
      values[key] = value
    }

    // form attrs
    const attrs = {
      novalidate: 'novalidate',
      onsubmit: this.handleSubmit.bind(this)
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

    const backButton = () => {
      if (this.local.machine.state.machine === 'basicInfo') return

      const attrs = {
        class: 'bg-white near-black dib bn b pv3 ph5 flex-shrink-0 f5 grow',
        style: 'outline:solid 1px var(--near-black);outline-offset:-1px',
        onclick: (e) => {
          e.preventDefault()

          this.local.machine.emit('machine:prev')
        }
      }

      return html`
        <button ${attrs}>
          ${icon('arrow')}
        </button>
      `
    }

    return html`
      <div class="base-form flex flex-column">
        <div>
          ${backButton()}
        </div>
        ${messages(this.state, this.local.form)}
        <h2 class="lh-title f3 fw1 mb5">${title}</h2>
        <div>
          ${this.renderProfileSwitcher.bind(this)()}
        </div>
        <div>
          ${this.renderRoleSwitcher.bind(this)()}
        </div>
        <form ${attrs}>
          ${Object.entries(elements)
            .map(([name, el]) => {
              // possibility to filter by name
              return el(this.validator, this.local.form)
            })}

          ${submitButton()}
        </form>
      </div>
    `
  }

  /**
   * Basic info form submit handler
   */
  handleSubmit (e) {
    e.preventDefault()

    if (!this.local.form.changed) {
      if (this.local.usergroup.groupType === 'label') {
        return this.local.machine.emit('machine:next')
      }
      return this.local.machine.emit('machine:end')
    }

    this.local.machine.emit('form:submit')
  }

  /**
   * Basic info load handler
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
   * Basic info form update handler
   * @returns {Boolean}
   */
  update (props = {}) {
    return props.role !== this.local.role
  }
}

module.exports = ProfileForm
