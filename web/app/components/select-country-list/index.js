/* global fetch */

const html = require('choo/html')
const Component = require('choo/component')
const { getData, getCountry: getName, getCode } = require('iso-3166-1-alpha-2') // migrating from gravityforms, we need to use alpha-2
const countryList = getData() // get country list data

class SelectCountryList extends Component {
  constructor (id, state, emit) {
    super(id)

    this.emit = emit
    this.state = state

    this.local = state.components[id] = {}

    this.local.options = countryList.map(({ code, name }) => {
      return {
        value: code,
        label: name
      }
    }).sort((a, b) => a.label.localeCompare(b.label, 'en', {}))

    this.onchange = this.onchange.bind(this)
  }

  createElement (props) {
    this.local.country = props.country

    return html`
      <div class="flex flex-column">
        <label for="country" class="f6 b db mr2">Select a country</label>
        <select id="country" class="ba bw b--gray bg-white black pa2" onchange=${this.onchange} name="country">
          <option value="" selected=${!this.local.country} disabled>…</option>
          ${this.local.options.map(({ value, label, disabled = false }) => {
            return html`
              <option value=${value} disabled=${disabled} selected=${getCode(this.local.country) === value}>
                ${label}
              </option>
            `
          })}
        </select>
      </div>
    `
  }

  async onchange (e) {
    const value = e.target.value

    this.local.country = getName(value)

    try {
      let response = await fetch('')

      const csrfToken = response.headers.get('X-CSRF-Token')

      response = await fetch('', {
        method: 'PUT',
        credentials: 'include',
        headers: {
          Accept: 'application/json',
          'X-CSRF-Token': csrfToken
        },
        body: new URLSearchParams({
          country: this.local.country
        })
      })

      this.state.profile.country = this.local.country
    } catch (err) {
      console.log(err)
    }
  }

  update (props) {
    return props.country !== this.local.country
  }
}

module.exports = SelectCountryList
