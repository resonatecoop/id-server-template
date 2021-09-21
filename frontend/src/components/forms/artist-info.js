const html = require('choo/html')
const Component = require('choo/component')
const input = require('@resonate/input-element')
const button = require('@resonate/button')
const isEmpty = require('validator/lib/isEmpty')
const messages = require('./messages')

const PaymentMethods = require('../payment-methods')
const TagsInput = require('../tags-input')
const AutocompleteInput = require('../autocomplete-typeahead')
const Dialog = require('@resonate/dialog-component')
const morph = require('nanomorph')

const validateFormdata = require('validate-formdata')

// ArtistInfoForm component class
class ArtistInfoForm extends Component {
  /***
   * Create an artist info form component
   * @param {String} id - The artist info form component id (unique)
   * @param {Number} state - The choo app state
   * @param {Function} emit - Emit event on choo app
   */
  constructor (id, state, emit) {
    super(id)

    this.emit = emit
    this.state = state

    this.local = state.components[id] = {}

    this.validator = validateFormdata()
    this.form = this.validator.state

    this.handleSubmit = this.handleSubmit.bind(this)
  }

  createElement (props) {
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
          <div class="mb5">
            <label for="music-genre" class="f4 db mv2">Genre</label>
            <p>Help others discover your music.</p>
            ${this.state.cache(TagsInput, 'tags-input').render({
              form: this.form,
              validator: this.validator,
              items: ['electro']
            })}
          </div>

          <div class="mb5">
            <label for="other-members" class="f4 db mv2">Other Members</label>

            <p>If you’re a band or group, add members to your profile.</p>

            ${this.state.cache(AutocompleteInput, 'autocomplete-input').render({
              form: this.form,
              validator: this.validator,
              title: 'Current members',
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

                    const content = function () {
                      const pristine = this.form.pristine
                      const errors = this.form.errors
                      const values = this.form.values
                      return html`
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
                                this.validator.validate(e.target.name, e.target.value)
                                morph(this.element.querySelector('.content'), this.content())
                              }
                            })}
                            <p class="ma0 pa0 message warning">
                              ${errors.displayName && !pristine.displayName ? errors.displayName.message : ''}
                            </p>
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
                                this.validator.validate(e.target.name, e.target.value)
                                morph(this.element.querySelector('.content'), this.content())
                              }
                            })}
                            <p class="ma0 pa0 message warning">
                              ${errors.role && !pristine.role ? errors.role.message : ''}
                            </p>
                          </div>

                          <div class="flex">
                            ${button({ type: 'submit', text: 'Continue' })}
                          </div>
                      </div>
                    `
                    }

                    const dialogEl = this.state.cache(Dialog, 'member-role').render({
                      title: 'Set member display name and role',
                      classList: 'dialog-default dialog--sm',
                      form,
                      validator,
                      onSubmit: function (e) {
                        e.preventDefault()

                        for (const field of e.target.elements) {
                          const isRequired = field.required
                          const name = field.name || ''
                          const value = field.value || ''

                          if (isRequired) {
                            this.validator.validate(name, value)
                          }
                        }

                        morph(this.element.querySelector('.content'), this.content())

                        if (this.form.valid) {
                          this.close()
                        }
                      },
                      content
                    })
                    document.body.appendChild(dialogEl)
                  }}>
                    ${item}
                  </div>
                `
              },
              placeholder: 'Members name',
              items: ['Artist 3', 'Artist 2', 'Artist 1']
            })}
          </div>

          <div class="mb5">
            <label for="payment-methods" class="f4 db mv2">Payment methods</label>

            <p>This is where we’ll send your earnings.</p>

            ${this.state.cache(PaymentMethods, 'payment-methods').render({
              form: this.form,
              validator: this.validator
            })}
          </div>

          <div class="mb5">
            <label for="recommended-artists" class="f4 db mv2">Recommended Artists <small class="f5">Optional</small></label>

            <p>Help others discover artists you’re associated with or inspired by. These names will appear on your profile.</p>

            <div class="flex">
              <div class="flex-auto w-100">
              ${this.state.cache(AutocompleteInput, 'autocomplete-input-2').render({
                form: this.form,
                validator: this.validator,
                title: 'Current artists',
                placeholder: 'Artists name',
                eachItem: function (item, index) {
                  return html`
                    <div>
                      ${item}
                    </div>
                  `
                },
                items: ['Artist 1', 'Artist 2']
              })}
              </div>
              <p class="ma0 ph2">
                Not on Resonate?
                <a class="link" onclick=${(e) => {}}>Invite them</a> now
              </p>
            </div>
          </div>

          ${submitButton()}
        </form>
      </div>
    `
  }

  handleSubmit (e) {
    e.preventDefault()

    for (const field of e.target.elements) {
      const isRequired = field.required
      const name = field.name || ''
      const value = field.value || ''
      if (isRequired) {
        this.validator.validate(name, value)
      }
    }

    this.rerender()
    const invalidInput = document.querySelector('.invalid')
    if (invalidInput) invalidInput.focus({ preventScroll: false }) // focus to first invalid input

    if (this.form.valid) {
      this.emit(this.state.events.PUSHSTATE, '/')
    }
  }

  update () {
    return true
  }
}

module.exports = ArtistInfoForm
