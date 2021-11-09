const Nanocomponent = require('nanocomponent')
const html = require('nanohtml')
const icon = require('@resonate/icon-element')
const Dialog = require('@resonate/dialog-component')
const button = require('@resonate/button')
const nanostate = require('nanostate')
const { getAPIServiceClientWithAuth } = require('@resonate/api-service')({
  apiHost: process.env.APP_HOST
})
const imagePlaceholder = require('../../lib/image-placeholder')

// UserMenu class
class UserMenu extends Nanocomponent {
  /***
   * Create user menu component
   * @param {String} id - The user menu component id (unique)
   * @param {Number} state - The choo app state
   * @param {Function} emit - Emit event on choo app
   */
  constructor (id, state, emit) {
    super(id)

    this.emit = emit
    this.state = state

    this.local = state.components[id] = {}

    this.local.machine = nanostate.parallel({
      creditsDialog: nanostate('close', {
        open: { close: 'close' },
        close: { open: 'open' }
      }),
      logoutDialog: nanostate('close', {
        open: { close: 'close' },
        close: { open: 'open' }
      })
    })

    this.local.src = imagePlaceholder(400, 400)

    this.local.machine.on('creditsDialog:open', async () => {
      // do something, redirects or open dialog
    })

    this.local.machine.on('logoutDialog:open', () => {
      const confirmButton = button({
        type: 'submit',
        value: 'yes',
        outline: true,
        theme: 'light',
        text: 'Log out'
      })

      const cancelButton = button({
        type: 'submit',
        value: 'no',
        outline: true,
        theme: 'light',
        text: 'Cancel'
      })

      const machine = this.local.machine

      const dialogEl = this.state.cache(Dialog, 'header-dialog').render({
        title: 'Log out',
        prefix: 'dialog-default dialog--sm',
        content: html`
          <div class="flex flex-column">
            <p class="lh-copy f5">Confirm you want to log out.</p>
            <div class="flex items-center">
              <div class="mr3">
                ${confirmButton}
              </div>
              <div>
                ${cancelButton}
              </div>
            </div>
          </div>
        `,
        onClose: function (e) {
          if (this.element.returnValue === 'yes') {
            // emit('logout', false)
            window.location.href = '/web/logout'
          }

          machine.emit('logoutDialog:close')
          this.destroy()
        }
      })

      document.body.appendChild(dialogEl)
    })
  }

  createElement () {
    return html`
      <li id="usermenu" role="menuitem" class="flex flex-auto justify-center w-100 mw4">
        <button title="Open menu" class="bg-transparent bn dropdown-toggle w-100 pa2 grow">
          <span class="flex justify-center items-center">
            <div class="fl w-100 mw2">
              <div class="db aspect-ratio aspect-ratio--1x1 bg-dark-gray bg-dark-gray--dark">
                <figure class="ma0">
                  <img src=${this.local.src} width="60" height="60" class="aspect-ratio--object z-1">
                  <figcaption class="clip"></figcaption>
                </figure>
              </div>
            </div>
            <div class="ph2">
              ${icon('caret-down', { size: 'xs' })}
            </div>
          </span>
        </button>
        <ul style="width:100vw;left:auto;max-width:18rem;margin-top:-1px;" role="menu" class="bg-white black bg-black--dark white--dark bg-white--light black--light ba bw b--mid-gray b--mid-gray--light b--near-black--dark list ma0 pa0 absolute right-0 dropdown z-999 bottom-100 top-100-l">
          <li role="menuitem" class="pt3">
            <div class="flex flex-auto items-center ph3">
              <span class="b">${this.local.displayName}</span>
            </div>
          </li>
          <li class="bb bw b--mid-gray b--mid-gray--light b--near-black--dark mv3" role="separator"></li>
          <li class="flex items-center ph3" role="menuitem">
            <div class="flex flex-column">
              <label for="credits">Credits</label>
              <input disabled tabindex="-1" name="credits" type="number" value=${this.local.credits} readonly class="bn br0 bg-transparent b ${this.local.credits < 0.128 ? 'red' : ''}">
            </Div>
            <div class="flex flex-auto justify-end">
              <button type="button" onclick=${(e) => { e.preventDefault(); this.local.machine.emit('creditsDialog:open') }} style="outline:solid 1px var(--near-black);outline-offset:-1px" class="pv2 ph3 ttu near-black near-black--light near-white--dark bg-transparent bn bn b flex-shrink-0 f6 grow">Add credits</button>
            </div>
          </li>
          <li class="bb bw b--mid-gray b--mid-gray--light b--near-black--dark mt3 mb2" role="separator"></li>
          <li class="mb1" role="menuitem">
            <a class="link db pv2 pl3" href="${process.env.APP_HOST}/faq">FAQ</a>
          </li>
          <li class="mb1" role="menuitem">
            <a class="link db pv2 pl3" target="blank" rel="noreferer noopener" href="https://resonate.is/support">Support</a>
          </li>
          <li class="mb1" role="menuitem">
            <a class="link db pv2 pl3" href="${process.env.APP_HOST}/settings">Settings</a>
          </li>
          <li class="bb bw b--mid-gray b--mid-gray--light b--near-black--dark mb3" role="separator"></li>
            <li class="pr3 pb3" role="menuitem">
              <div class="flex justify-end">
                ${button({
                  prefix: 'ttu near-black near-black--light near-white--dark f6 ba b--mid-gray b--mid-gray--light b--dark-gray--dark',
                  onClick: (e) => this.local.machine.emit('logoutDialog:open'),
                  style: 'blank',
                  text: 'Log out',
                  outline: true
                })}
              </div>
            </li>
        </ul>
      </li>
    `
  }

  async load (el) {
    try {
      // get v2 api profile
      const getClient = getAPIServiceClientWithAuth(this.state.token)
      const client = await getClient('profile')
      const result = await client.getUserProfile()

      const { body: response } = result
      const { data: userdata } = response

      this.local.src = userdata.avatar['profile_photo-m'] || userdata.avatar['profile_photo-l'] || imagePlaceholder(400, 400)

      this.local.credits = userdata.credits

      this.local.displayName = userdata.nickname

      this.rerender()

      console.log(response)
    } catch (err) {
      console.log(err.message)
      console.log(err)
    }
  }
}

module.exports = UserMenu
