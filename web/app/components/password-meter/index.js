const html = require('choo/html')
const Component = require('choo/component')
const zxcvbnAsync = require('zxcvbn-async')

class PasswordMeter extends Component {
  constructor (id, state, emit) {
    super(id)

    this.emit = emit
    this.state = state

    this.local = state.components[id] = {}
  }

  createElement (props) {
    this.local.password = props.password

    const score = this.getScore(this.local.password)
    const meter = [0, 1, 2, 3, 4] // zxcvbn score meter

    return html`
      <div class="flex flex-column">
        <div class="flex">
        ${this.local.password ? meter.map((n) => {
          const colors = [
            'dark-red',
            'red',
            'orange',
            'green',
            'dark-green'
          ]
          const color = n <= score ? colors[score] : 'gray'

          return html`
            <div style="height:3px" class="flex-auto w-100 bg-${color}"></div>
          `
        }) : ''}
        </div>
      </div>
    `
  }

  getScore (password) {
    if (!password) return 0

    const zxcvbn = zxcvbnAsync.load({
      sync: true,
      libUrl: 'https://cdn.jsdelivr.net/npm/zxcvbn@4.4.2/dist/zxcvbn.js',
      libIntegrity: 'sha256-9CxlH0BQastrZiSQ8zjdR6WVHTMSA5xKuP5QkEhPNRo='
    })
    const { score } = zxcvbn(this.local.password || '')
    return score
  }

  update (props) {
    return props.password !== this.local.password
  }
}

module.exports = PasswordMeter
