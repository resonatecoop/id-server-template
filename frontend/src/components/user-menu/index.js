const Nanocomponent = require('nanocomponent')
const html = require('nanohtml')
const Dialog = require('@resonate/dialog-component')
const button = require('@resonate/button')
const nanostate = require('nanostate')
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

    this.local.usergroup = {
      avatar: imagePlaceholder(400, 400)
    }

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
        title: 'Logout from Resonate',
        prefix: 'dialog-default dialog--sm',
        content: html`
          <div class="flex flex-column">
            <p class="lh-copy f5">Please confirm the action.</p>
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
            window.location.href = '/web/logout'
          }

          machine.emit('logoutDialog:close')
          this.destroy()
        }
      })

      document.body.appendChild(dialogEl)
    })
  }

  createElement (props = {}) {
    this.local.displayName = props.displayName

    return html`
      <ul id="usermenu" style="width:100vw;left:auto;max-width:18rem;margin-top:-1px;" role="menu" class="bg-white black bg-black--dark white--dark bg-white--light black--light ba bw b--mid-gray b--mid-gray--light b--near-black--dark list ma0 pa0 absolute right-0 dropdown z-999 bottom-100 top-100-l">
        <li role="menuitem" class="pt3">
          <div class="flex flex-auto items-center ph3">
            <span class="b">${this.local.displayName}</span>
          </div>
        </li>
        <li class="bb bw b--mid-gray b--mid-gray--light b--near-black--dark mv3" role="separator"></li>
        <li class="flex items-center ph3" role="menuitem">
          <div class="flex flex-column">
            <label for="credits">Credits</label>
            <input disabled tabindex="-1" name="credits" type="number" value=${this.state.profile.credits} readonly class="bn br0 bg-transparent b ${this.state.profile.credits < 0.128 ? 'red' : ''}">
          </Div>
          <div class="flex flex-auto justify-end">
          </div>
        </li>
        <li class="bb bw b--mid-gray b--mid-gray--light b--near-black--dark mt3 mb2" role="separator"></li>
        <li class="mb1" role="menuitem">
          <a class="link db pv2 pl3" href="/account">Update your account</a>
        </li>
        <li class="mb1" role="menuitem">
          <a class="link db pv2 pl3" href="/account-settings">Account settings</a>
        </li>
        <li class="mb1" role="menuitem">
          <a class="link db pv2 pl3" href="${process.env.APP_HOST}/faq">FAQ</a>
        </li>
        <li class="mb1" role="menuitem">
          <a class="link db pv2 pl3" target="blank" href="https://resonate.is/support">Support</a>
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
    `
  }
}

module.exports = UserMenu
