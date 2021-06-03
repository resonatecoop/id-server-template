/* global fetch */

const Component = require('choo/component')
const html = require('choo/html')
const icon = require('@resonate/icon-element')

// ProfileTypeForm component class
class ProfileTypeForm extends Component {
  /***
   * Create profile type form component
   * @param {String} id - The account type form component id (unique)
   * @param {Number} state - The choo app state
   * @param {Function} emit - Emit event on choo app
   */
  constructor (id, state, emit) {
    super(id)

    this.emit = emit
    this.state = state

    this.local = state.components[id] = {}

    this.local.value = 'listener'

    this.local.items = [
      { value: 'listener', name: 'Listener' },
      { value: 'artist', name: 'Artist' },
      { value: 'label', name: 'Label' }
    ]

    this.handleKeyPress = this.handleKeyPress.bind(this)
    this.updateSelection = this.updateSelection.bind(this)
    this.handleSubmit = this.handleSubmit.bind(this)
  }

  /***
   * Create profile type form component element
   * @param {Object} props - The account type form component props
   * @param {String} props.value - Account type to select (defaults to listener)
   * @returns {HTMLElement}
   */
  createElement (props = {}) {
    if (!this.local.value) {
      this.local.value = props.value
    }

    // form attrs
    const attrs = {
      novalidate: 'novalidate',
      class: 'flex flex-column flex-row-l',
      onsubmit: this.handleSubmit
    }

    const submitButton = () => {
      // button attrs
      const attrs = {
        type: 'submit',
        class: 'bg-white near-black dib bn b pv3 ph5 flex-shrink-0 f5 grow',
        style: 'outline:solid 1px var(--near-black);outline-offset:-1px',
        text: 'Continue'
      }
      return html`
        <button ${attrs}>
          Continue
        </button>
      `
    }

    // input radio label
    const label = (name, id) => {
      // label attrs
      const attrs = {
        class: 'flex items-center fw4 pv3 w-100 grow bw',
        style: 'outline:solid 1px var(--near-black);outline-offset:-1px',
        tabindex: '0',
        onkeypress: this.handleKeyPress,
        for: id
      }
      return html`
        <label ${attrs}>
          <div class="flex flex-shrink-0 justify-center bg-white items-center w2 h2 ml2">
            ${icon('check', { size: 'sm', class: 'fill-transparent' })}
          </div>
          ${name}
        </label>
      `
    }

    return html`
      <form ${attrs}>
        <div class="flex flex-column flex-auto w-50 mb3">
          ${this.local.items.map((item, index) => {
            // input attrs
            const attrs = {
              onchange: this.updateSelection,
              id: 'item-' + index,
              tabindex: -1,
              name: 'item',
              type: 'radio',
              checked: item.value === this.local.value,
              value: item.value
            }

            return html`
              <div>
                <input ${attrs}>
                ${label(item.name, 'item-' + index)}
              </div>
            `
          })}
        </div>
        <div class="flex flex-auto flex-column ph4-l w-100">
          ${this.renderInfo()}
          <div class="flex justify-end">
            ${submitButton()}
          </div>
        </div>
      </form>
    `
  }

  renderInfo () {
    return {
      listener: html`
        <dl>
          <dt class="f3 lh-title fw1">Listener Account</dt>
          <dd class="ma0">
            <p>Your Listener Account comes with .128 Resonate credits (about 4 hours of listening time). Login to the <a class="b" href="https://beta.stream.resonate.coop/api/v2/user/connect/resonate" target>Resonate Player</a> to buy credit to top-up your stream2own balance.</p>
            <p>Consider becoming a member of the Resonate co-op for only 5€ a year.</p>
          </p>
        </dl>
      `,
      artist: html`
        <dl>
          <dt class="f3 lh-title fw1">Artist Account</dt>
          <dd class="ma0">
            <p>Own your streaming platform. Join an active co-op with your music to get paid fairly and transparently. Govern the development and direction of Resonate. With an Artist Account, you can upload your music for stream2own on the <a class="b" href="https://beta.stream.resonate.coop/api/v2/user/connect/resonate" target>Resonate Player</a>. An Artist Account is for managing your own music, or on behalf of a band or artist you represent.</p>
            <p>Artists earn membership to the co-op by uploading music.</p>
          </p>
        </dl>
      `,
      label: html`
        <dl>
          <dt class="f3 lh-title fw1">Label Account</dt>
          <dd class="ma0">
            <p>We are currently building a new catalog processing infrastructure for Resonate. Until this project is complete, new signups have been paused for Label Accounts. Email <a class="b" href="mailto:members@resonate.is">members@resonate.is</a> to be notified when signup is live again. Thank you for your support.</p>
            <p>With a Label Account you will be able to create and manage accounts for your artists. Membership: As a Label, earn membership in the co-op by uploading music to Resonate on behalf of your artists. Membership is automatically earned for artists who have had their music uploaded by their label.</p>
          </p>
        </dl>
      `
    }[this.local.value]
  }

  updateSelection (e) {
    const val = e.target.value
    this.local.value = val
    this.rerender()
  }

  handleKeyPress (e) {
    if (e.keyCode === 13) {
      e.preventDefault()
      e.target.control.checked = !e.target.control.checked
      const val = e.target.control.value
      this.local.value = val
      this.rerender()
    }
  }

  async handleSubmit (e) {
    e.preventDefault()

    try {
      let response = await fetch('')

      const csrfToken = response.headers.get('X-CSRF-Token')

      const role = {
        artist: 'member',
        listener: 'fans',
        label: 'label-owner'
      }[this.local.value]

      const payload = { role }

      response = await fetch('', {
        method: 'PUT',
        credentials: 'include',
        headers: {
          Accept: 'application/json',
          'X-CSRF-Token': csrfToken
        },
        body: new URLSearchParams(payload)
      })

      this.emit('set:usergroup', this.local.value) // set usergroup (wp role)
    } catch (err) {
      console.log(err)
    }
  }

  update (props) {
    return props.value !== this.local.value
  }
}

module.exports = ProfileTypeForm
