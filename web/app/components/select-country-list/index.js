const html = require('choo/html')
const Component = require('choo/component')
const { getData, getCode } = require('country-list')
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
      <div class="flex flex-column ml2 mb3">
        <label for="country" class="f6 b db mr2">Select a country</label>
        <select id="country" class="ba bw b--gray bg-white black" onchange=${this.onchange} name="country">
          ${this.local.options.map(({ value, label, disabled = false }) => {
            return html`
              <option value=${value} disabled=${disabled} selected=${this.local.country === value}>
                ${label}
              </option>
            `
          })}
        </select>
      </div>
    `
  }

  onchange (e) {
    const value = e.target.value
    this.local.country = value
    this.local.code = getCode(value)
    // TODO send request to update country
  }

  update (props) {
    return props.country !== this.local.country
  }
}

module.exports = SelectCountryList
