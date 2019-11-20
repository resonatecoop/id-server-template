const html = require('choo/html')
const Component = require('choo/component')
const nanostate = require('nanostate')
const icon = require('@resonate/icon-element')
const button = require('@resonate/button')

/* global fetch */

const lifetimes = [
  { value: 3600, text: '1 hour', id: 'hour' },
  { value: 86400, text: '1 day', id: 'day' },
  { value: 604800, text: '1 week', id: 'week' }
]

class Authorize extends Component {
  constructor (id, state, emit) {
    super(id)

    this.emit = emit
    this.state = state

    this.local = state.components[id] = {}

    this.local.data = {
      allow: 'Allow'
    }

    this.machine = nanostate('idle', {
      idle: { start: 'loading' },
      loading: { resolve: 'idle', reject: 'error' },
      error: { start: 'loading' }
    })

    this.machine.on('loading', () => {
      this.rerender()
    })

    this.machine.on('error', () => {
      this.rerender()
    })
  }

  createElement (props) {
    return html`
      <div class="flex flex-column flex-auto">
        <form action="" method="post">
          <div class="flex flex-column">
            <p><b>${this.state.query.client_id}</b> would like to perform actions on your behalf.</p>

            <p>How long do you want to authorize <b>${this.state.query.client_id}</b> for?</p>

            <div class="flex flex-auto w-100">
              ${lifetimes.map(({ value, text, id }) => html`
                <div class="flex items-center flex-auto">
                  <input
                    type=radio
                    name="lifetime"
                    id=${id}
                    checked=${this.local.data.lifetime === Number(value)}
                    value=${value}
                    class="o-0"
                    style="width:0;height:0;"
                    onchange=${(e) => {
                      this.local.data.lifetime = Number(e.target.value)
                      this.rerender()
                    }}>

                  <label class="flex justify-center items-center w-100 dim" for=${id}>
                    <div class="flex w-100 flex-auto">
                      <div class="flex items-center justify-center h1 w1 ba bw b--mid-gray">
                        ${icon('circle', { class: `icon icon--xxs ${this.local.data.lifetime === Number(value) ? 'fill-black fill-white--dark fill-black--light' : 'fill-transparent'}` })}
                      </div>
                      <div class="flex items-center ph3 f5 lh-copy">
                        ${text}
                      </div>
                    </div>
                  </label>
                </div>
              `)}
            </div>
          </div>
          <div class="flex">
            <div class="mr2">
              ${button({
                type: 'submit',
                name: 'allow',
                prefix: 'bg-white ba bw b--dark-gray f5 b pv3 ph5 grow',
                value: 'Allow',
                text: 'Allow',
                style: 'none',
                size: 'none'
              })}
            </div>
            <div>
              ${button({
                type: 'submit',
                name: 'deny',
                prefix: 'bg-white ba bw b--dark-gray f5 b pv3 ph5 grow',
                value: 'Deny',
                text: 'Deny',
                style: 'none',
                size: 'none'
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
