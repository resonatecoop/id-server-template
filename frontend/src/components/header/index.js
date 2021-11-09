const Nanocomponent = require('nanocomponent')
const nanostate = require('nanostate')
const html = require('nanohtml')
const icon = require('@resonate/icon-element')
const Search = require('@resonate/search-component')

// SearchOuter class
// base for search component we use on Resonate website and id server
class SearchOuter extends Nanocomponent {
  /***
   * Create a search outer component
   * @param {String} id - The search outer component id (unique)
   * @param {Number} state - The choo app state
   * @param {Function} emit - Emit event on choo app
   */
  constructor (id, state, emit) {
    super(id)

    this.state = state
    this.emit = emit
    this.local = state.components[id] = {}

    this.local.tags = [
      'ambient',
      'acoustic',
      'alternative',
      'electro',
      'electronic',
      'experimental',
      'folk',
      'funk',
      'hiphop',
      'house',
      'indie',
      'instrumental',
      'jazz',
      'metal',
      'pop',
      'punk'
    ]

    this.local.machine = nanostate.parallel({
      search: nanostate('off', {
        on: { toggle: 'off' },
        off: { toggle: 'on' }
      })
    })

    this.local.machine.on('search:toggle', () => {
      this.rerender()
      if (this.local.machine.state.search === 'on') {
        const input = document.querySelector('input[type="search"]')
        if (input && input !== document.activeElement) input.focus()
      }
      document.body.classList.toggle('search-open', this.local.machine.state.search === 'on')
    })
  }

  createElement () {
    const machine = {
      on: () => {
        return this.state.cache(Search, 'search').render({
          tags: this.local.tags,
          applicationHost: process.env.APP_HOST
        })
      },
      off: () => {
        const attrs = {
          onclick: (e) => {
            this.local.machine.emit('search:toggle')
          },
          class: 'bn mr4 mr0-l pa0 bg-transparent'
        }
        return html`
          <button ${attrs}>
            <div class="flex items-center justify-center">
              ${icon('search', { size: 'sm' })}
              <span class="dn db-l pl3 near-black near-black--light near-white--dark">Search</span>
            </div>
          </button>
        `
      }
    }[this.local.machine.state.search]

    return html`
      <li id="search-host" role="menuitem" class="flex w-100 flex-auto justify-end justify-center-l">
        ${machine()}
      </li>
    `
  }

  update () {
    return false
  }
}

module.exports = SearchOuter
