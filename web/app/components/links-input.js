const html = require('choo/html')
const Component = require('choo/component')
const validateFormdata = require('validate-formdata')
const input = require('@resonate/input-element')
const isURL = require('validator/lib/isURL')
const button = require('@resonate/button')

// LinksInput class
class LinksInput extends Component {
  constructor (id, state, emit) {
    super(id)

    this.emit = emit
    this.state = state

    this.local = state.components[id] = {}

    this.removeLink = this.removeLink.bind(this)
    this.addLink = this.addLink.bind(this)

    this.local.links = []

    this.validator = validateFormdata()
    this.form = this.validator.state

    this._onchange = this._onchange.bind(this)
  }

  createElement (props) {
    this.onchange = typeof props.onchange === 'function' ? props.onchange.bind(this) : this._onchange
    this.validator = props.validator || this.validator
    this.form = props.form || this.form || {
      changed: false,
      valid: true,
      pristine: {},
      required: {},
      values: {},
      errors: {}
    }

    this.local.required = props.required || false

    const pristine = this.form.pristine
    const errors = this.form.errors
    const values = this.form.values

    return html`
      <div class="flex flex-column">
        <p>Website, Instagram, Twitter, Mastodon, etc.</p>
        <div class="flex items-center mb1">
          ${input({
            type: 'url',
            name: 'link',
            invalid: errors.link && !pristine.link,
            required: false,
            placeholder: 'URL',
            value: values.link,
            onchange: (e) => {
              this.validator.validate(e.target.name, e.target.value)
              this.rerender()
            }
          })}
          ${button({
            onclick: (e) => this.addLink(values.link, errors.link),
            prefix: 'db bg-white bw b--black-20 h-100 ml1 pa3 grow',
            style: 'none',
            size: 'none',
            iconName: 'add',
            iconSize: 'sm'
          })}
          ${!this.local.required ? html`<p class="lh-copy f5 pl2 grey">Optional</p>` : ''}
        </div>

        <p class="ma0 pa0 message warning">
          ${errors.link && !pristine.link ? errors.link.message : ''}
        </p>

        <ul class="list ma0 pa0">
          ${this.local.links.map((link, index) => {
            return html`
              <li class="flex items-center relative">
                ${link}
                ${button({
                  onclick: (e) => this.removeLink(index),
                  prefix: 'bg-transparent bn ml2 pa0 grow',
                  iconName: 'close-fat',
                  iconSize: 'xxs',
                  style: 'none',
                  size: 'none'
                })}
              </li>
            `
          })}
        </ul>
      </div>
    `
  }

  _onchange () {}

  removeLink (index) {
    if (index > -1) {
      this.local.links.splice(index, 1)
      this.onchange(this.local.links)
      this.rerender()
    }
  }

  addLink (value, error) {
    if (value && !this.local.links.includes(value) && !error) {
      this.local.links.push(value)
      this.onchange(this.local.links)
      this.rerender()
    }
  }

  load () {
    this.validator.field('link', { required: this.local.required }, (data) => {
      if (!isURL(data, { require_protocol: false })) { return new Error('Link is not valid url') }
    })
  }

  update () {
    return false
  }
}

module.exports = LinksInput
