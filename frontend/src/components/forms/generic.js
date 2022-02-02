const html = require('choo/html')
const Component = require('choo/component')
const icon = require('@resonate/icon-element')
const input = require('@resonate/input-element')

class Form extends Component {
  constructor (id, state, emit) {
    super(id)

    this.emit = emit
    this.state = state

    this.local = state.components[id] = {}
    this.local.submitted = false
  }

  createElement (props) {
    this.form = props.form
    this.validate = props.validate
    this.submit = props.submit

    this.local.fields = props.fields || []
    this.local.altButton = props.altButton
    this.local.buttonText = props.buttonText || ''
    this.local.id = props.id
    this.local.action = props.action
    this.local.method = props.method || 'POST'

    const pristine = this.form.pristine
    const errors = this.form.errors
    const values = this.form.values

    const inputs = this.local.fields.map(fieldProps => {
      if (fieldProps.component) return fieldProps.component

      const { name = fieldProps.type, help, component } = fieldProps

      fieldProps.onInput = typeof fieldProps.onInput === 'function'
        ? fieldProps.onInput.bind(this)
        : null

      const element = component || input(Object.assign({}, fieldProps, {
        onchange: (e) => {
          this.validate({
            name: e.target.name,
            value: e.target.value
          })
          this.rerender()
        },
        value: values[name]
      }))

      return html`
        <div class="mb3">
          ${fieldProps.label
            ? html`
                <label for=${fieldProps.id || name} class="f5 db mb1 dark-gray">
                  ${fieldProps.label}
                </label>
              `
            : ''
          }
          <div class="relative">
            ${element}
            ${errors[name] && !pristine[name]
              ? html`
                <div class="absolute left-0 ph1 flex items-center" style="top:50%;transform: translate(-100%, -50%);">
                  ${icon('info', { class: 'fill-red', size: 'sm' })}
                </div>
              `
              : ''
            }
          </div>
          ${typeof help === 'function' ? help(values[name]) : help}
          ${errors[name] && !pristine[name]
            ? html`<span class="message f5 pb2">${errors[name].message}</span>`
            : ''
          }
        </div>
      `
    })

    // form attributes
    const attrs = {
      novalidate: 'novalidate',
      class: 'flex flex-column flex-auto',
      id: this.local.id,
      action: this.local.action,
      method: this.local.method,
      onsubmit: (e) => {
        e.preventDefault()

        for (const field of e.target.elements) {
          const isRequired = field.required
          const name = field.name || ''
          const value = field.value || ''
          if (isRequired) this.validate({ name, value })
        }

        if (this.form.valid) {
          this.submit(e.target)
        }

        this.rerender()
      }
    }

    const submitButton = (props = {}) => {
      const attrs = Object.assign({
        disabled: false,
        class: `bg-white dib bn pv3 ph5 flex-shrink-0 f5 ${props.disabled ? 'o-50' : 'grow'}`,
        style: 'outline:solid 1px var(--near-black);outline-offset:-1px',
        type: 'submit'
      }, props)

      return html`<button ${attrs}>${this.local.buttonText}</button>`
    }

    return html`
      <div class="flex flex-column flex-auto">
        <form ${attrs}>
          ${inputs}
          ${this.local.altButton
            ? html`
              <div class="flex mt3">
                <div class="flex mr3">
                  ${this.local.altButton}
                </div>
                <div class="flex flex-auto justify-end">
                  ${submitButton({ disabled: this.local.submitted })}
                </div>
              </div>`
            : html`
              <div class="flex flex-auto">
                ${submitButton({ disabled: this.local.submitted })}
              </div>`
            }
        </form>
      </div>
    `
  }

  update () {
    return false
  }
}

module.exports = Form
