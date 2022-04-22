const html = require('choo/html')
const Component = require('choo/component')
const nanostate = require('nanostate')
const icon = require('@resonate/icon-element')

class Authorize extends Component {
  constructor (id, state, emit) {
    super(id)

    this.emit = emit
    this.state = state

    this.local = state.components[id] = Object.create({
      machine: nanostate.parallel({
        request: nanostate('idle', {
          idle: { start: 'loading' },
          loading: { resolve: 'data', reject: 'error', reset: 'idle' },
          data: { reset: 'idle', start: 'loading' },
          error: { reset: 'idle', start: 'loading' }
        }),
        loader: nanostate('off', {
          on: { toggle: 'off' },
          off: { toggle: 'on' }
        })
      })
    })

    this.local.error = {}

    this.local.machine.on('request:error', () => {
      if (this.element) this.rerender()
    })

    this.local.machine.on('request:loading', () => {
      if (this.element) this.rerender()
    })

    this.local.machine.on('loader:toggle', () => {
      if (this.element) this.rerender()
    })

    this.local.machine.on('request:error', () => {
      if (this.element) this.rerender()
    })

    this.local.machine.transitions.request.event('error', nanostate('error', {
      error: { start: 'loading' }
    }))

    this.local.machine.on('request:noResults', () => {
      if (this.element) this.rerender()
    })

    this.local.machine.transitions.request.event('noResults', nanostate('noResults', {
      noResults: { start: 'loading' }
    }))
  }

  createElement (props) {
    const input = (props) => {
      const attrs = Object.assign({
        type: 'submit',
        name: 'allow',
        class: 'bg-white black ba bw b--dark-gray f5 b pv3 ph3 grow',
        value: 'Finish Login'
      }, props)

      return html`
        <input ${attrs}>
      `
    }

    const attrs = {
      class: 'flex flex-column flex-auto ma0 pa0',
      action: '',
      method: 'POST'
    }

    return html`
      <div class="flex flex-column flex-auto">
        <form ${attrs}>
          <input type="hidden" name="gorilla.csrf.Token" value=${this.state.csrfToken}>

          ${this.state.query.response_type === 'token'
          ? html`
            <p>How long do you want to authorize <b>${this.state.applicationName}</b> for?</p>
            <div class="flex w-100">
              <div class="flex items-center flex-auto">
                <input tabindex="-1" type="radio" name="lifetime" id="hour" value="3600">
                <label tabindex="0" class="flex flex-auto items-center justify-center w-100" for="hour">
                  <div class="pv3 flex justify-center w-100 flex-auto">
                    ${icon('circle', { class: 'fill-white' })}
                  </div>
                  <div class="pv3 flex w-100 flex-auto">1 hour</div>
                </label>
              </div>
              <div class="flex items-center flex-auto">
                <input tabindex="-1" type="radio" name="lifetime" id="day" value="86400">
                <label tabindex="0" class="flex flex-auto items-center justify-center w-100" for="day">
                  <div class="pv3 flex justify-center w-100 flex-auto">
                    ${icon('circle', { class: 'fill-white' })}
                  </div>
                  <div class="pv3 flex w-100 flex-auto">1 day</div>
                </label>
              </div>
              <div class="flex items-center flex-auto">
                <input tabindex="-1" type="radio" name="lifetime" id="week" value="604800" checked>
                <label tabindex="0" class="flex flex-auto items-center justify-center w-100" for="week">
                  <div class="pv3 flex justify-center w-100 flex-auto">
                    ${icon('circle', { class: 'fill-white' })}
                  </div>
                  <div class="pv3 flex w-100 flex-auto">1 week</div>
                </label>
              </div>
            </div>`
          : ''}

          <p class="lh-copy">Logging in as <b>${this.state.profile.displayName ? this.state.profile.displayName : this.state.profile.email}</b></p>

          <div class="flex">
            <div class="mr3">
              ${input({
                name: 'allow',
                value: 'Finish Login'
              })}
            </div>
            <div>
              ${input({
                name: 'deny',
                class: 'bg-white black f5 bn b pv3 ph3 grow',
                value: 'Cancel'
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
