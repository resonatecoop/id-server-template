const html = require('choo/html')
const Component = require('choo/component')

class Authorize extends Component {
  constructor (id, state, emit) {
    super(id)

    this.emit = emit
    this.state = state

    this.local = state.components[id] = {}
  }

  createElement (props) {
    const input = (props) => {
      const attrs = Object.assign({
        type: 'submit',
        name: 'allow',
        class: 'bg-white ba bw b--dark-gray f5 b pv3 ph5 grow',
        value: 'Allow'
      }, props)

      return html`
        <input ${attrs}>
      `
    }

    return html`
      <div class="flex flex-column flex-auto">
        <form action="" method="post">
          <div class="flex flex-column">
            <p><b>${this.state.query.client_id}</b> would like to perform actions on your behalf.</p>
          </div>
          <div class="flex">
            <div class="mr2">
              ${input({
                name: 'allow',
                value: 'Allow'
              })}
            </div>
            <div>
              ${input({
                name: 'deny',
                value: 'Deny'
              })}
            </div>
          </div>
        </form>
      </div>
    `
  }

  update () {
    return false
  }
}

module.exports = Authorize
