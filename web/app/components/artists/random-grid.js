/* global fetch */

const Component = require('choo/component')
const html = require('choo/html')
const logger = require('nanologger')
const log = logger('artists-grid')

class ArtistsRandomGrid extends Component {
  constructor (name, state, emit) {
    super(name)

    this.name = name
    this.ids = []
    this.items = []

    this.state = state
    this.emit = emit

    this.fetch = this.fetch.bind(this)
  }

  createElement () {
    const item = ({ avatar, name: artist, id }) => {
      const filename = avatar.original || avatar.medium
      const url = filename || '/thumbs/default.png'
      return html`
        <li class="fl w-50 w-third-m w-20-l">
          <div class="db aspect-ratio aspect-ratio--1x1">
            <span role="img" aria-label=${artist} style="background: var(--near-black) url(${url}) no-repeat;" class="bg-center gray-100 hover-grayscale-0 cover aspect-ratio--object">
            </span>
          </div>
        </li>
      `
    }

    const items = this.items
      .filter(({ avatar }) => !!avatar)
      .slice(0, 24)
      .map(item)

    return html`
      <div class="fixed absolute--fill">
        <ul class="list ma0 pa0 cf">
          ${items}
        </ul>
      </div>
    `
  }

  async fetch () {
    try {
      const url = new URL('/v1/artists', 'https://api.resonate.is')
      url.search = new URLSearchParams({
        limit: 100,
        order: 'random'
      })

      const response = await (await fetch(url.href)).json()

      if (response.data) {
        this.items = response.data
        this.rerender()
      }
    } catch (err) {
      log.error(err)
    }
  }

  load () {
    if (!this.items.length) {
      this.fetch()
    }
  }

  update () {
    return false
  }
}

module.exports = ArtistsRandomGrid
