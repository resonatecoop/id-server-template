const html = require('choo/html')
const Component = require('choo/component')
const validateFormdata = require('validate-formdata')

const AutocompleteInput = require('../autocomplete-typeahead')

// LabelInfoForm component class
class LabelInfoForm extends Component {
  /***
   * Create a label info form component
   * @param {String} id - The label info form component id (unique)
   * @param {Number} state - The choo app state
   * @param {Function} emit - Emit event on choo app
   */
  constructor (id, state, emit) {
    super(id)

    this.emit = emit
    this.state = state

    this.handleSubmit = this.handleSubmit.bind(this)

    this.validator = validateFormdata()
    this.form = this.validator.state
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
        <form ${attrs}>
          <div class="mb5">
            <label for="your-artists" class="f4 db mv2">Your Artists</label>

            <p>Add the artists signed to your label. They’ll appear on your profile.</p>

            ${this.state.cache(AutocompleteInput, 'autocomplete-input-3').render({
              form: this.form,
              validator: this.validator,
              placeholder: 'Artist name',
              title: 'Current artists',
              items: ['Artist 1', 'Artist 2', 'Artist 3'],
              eachItem: function (item, index) {
                return html`
                  <div>${item}</div>
                `
              }
            })}
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
      this.emit(this.state.events.PUSHSTATE, '/account-settings') // TODO redirect to something else
    }
  }

  update () {
    return true
  }
}

module.exports = LabelInfoForm
