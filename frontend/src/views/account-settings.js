/* global fetch */

const html = require('choo/html')

const Dialog = require('@resonate/dialog-component')
const Button = require('@resonate/button-component')

const UpdateEmailForm = require('../components/forms/emailUpdate')
const UpdatePasswordForm = require('../components/forms/passwordUpdate')

const navigateToAnchor = (e) => {
  const el = document.getElementById(e.target.hash.substr(1))
  if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' })
  e.preventDefault()
}

/**
 * Account settings
 * @param {Object} state Choo state
 * @param {Function} emit Emit choo event (nanobus)
 */
module.exports = (state, emit) => {
  const deleteButton = new Button('delete-profile-button')

  return html`
    <div class="flex flex-column w-100 mh3 mh0-ns">
      <section id="account-settings" class="flex flex-column">
        <h2 class="lh-title pl3 f2 fw1">Account settings</h2>
        <div class="flex flex-column flex-row-l">
          <div class="w-50 w-third-l ph3">
            <nav class="sticky z-1 flex flex-column" style="top:3rem">
              <ul class="list ma0 pa0 mt3 flex flex-column">
                <li class="mb2">
                  <a class="link" href="#change-email" onclick=${navigateToAnchor}>Email</a>
                </li>
                <li class="mb2">
                  <a class="link" href="#change-password" onclick=${navigateToAnchor}>Password</a>
                </li>
                <li>
                  <a class="link" href="#delete-account" onclick=${navigateToAnchor}>Delete account</a>
                </li>
              </ul>
            </nav>
          </div>
          <div class="flex flex-column flex-auto ph3 pt4 pt0-l mw6 ph0-l">
            <section class="ph3 pb6">
              <h3 class="f3 fw1 lh-title relative mb4">
                Change email
                <a id="change-email" class="absolute" style="top:-120px"></a>
              </h3>
              ${state.cache(UpdateEmailForm, 'update-email').render({
                data: state.profile || {}
              })}
            </section>

            <section class="ph3">
              <h3 class="f3 fw1 lh-title relative mb4">
                Change password
                <a id="change-password" class="absolute" style="top:-120px"></a>
              </h3>
              ${state.cache(UpdatePasswordForm, 'update-password-form').render()}
            </section>

            <section class="flex w-100 items-center ph3">
              ${deleteButton.render({
                type: 'button',
                prefix: 'bg-white ba bw b--dark-gray f5 b pv3 ph3 w-100 mw5 grow',
                text: 'Delete account',
                style: 'none',
                onClick: () => {
                  const dialog = state.cache(Dialog, 'delete-account-dialog')
                  const dialogEl = dialog.render({
                    title: 'Delete account',
                    prefix: 'dialog-default dialog--sm',
                    onClose: async (e) => {
                      const { returnValue } = e.target
                      const { value: password } = e.target.querySelector('input[type=password]')

                      dialog.destroy()

                      if (returnValue === 'Delete account') {
                        try {
                          let response = await fetch('')

                          const csrfToken = response.headers.get('X-CSRF-Token')

                          response = await fetch('', {
                            method: 'PUT',
                            headers: {
                              Accept: 'application/json',
                              'X-CSRF-Token': csrfToken
                            },
                            body: new URLSearchParams({
                              password: password,
                              _method: 'DELETE'
                            })
                          })

                          if (response.status >= 400) {
                            const { error: errorMessage } = await response.json()

                            emit('notify', {
                              timeout: 10000,
                              type: 'info',
                              message: errorMessage
                            })
                          } else {
                            emit('notify', {
                              timeout: 3000,
                              type: 'info',
                              message: 'Your account has been scheduled for deletion in 24 hours. You will receive one last email to confirm or cancel the deletion.'
                            })

                            setTimeout(() => {
                              window.location.reload()
                            }, 3000)
                          }
                        } catch (err) {
                          console.log(err.message)
                          emit('notify', {
                            timeout: 10000,
                            type: 'info',
                            message: 'Account not deleted.'
                          })
                        }
                      }
                    },
                    content: html`
                      <div class="flex flex-column">
                        <p class="lh-copy f5 b">Are you sure you want to delete your Resonate account ?</p>

                        <div class="mb3">
                          <input autocomplete="off" class="bg-black white bg-white--dark black--dark bg-black--light white--light placeholder--dark-gray input-reset w-100 bn pa3 valid" type="password" placeholder="Password" required="required" name="password">
                        </div>

                        <div class="flex">
                          <div class="flex items-center">
                            <input class="bg-white black ba bw b--near-black f5 b pv2 ph3 ma0 grow" type="submit" value="Not really">
                          </div>
                          <div>
                          </div>
                          <div class="flex flex-auto w-100 justify-end">
                            <div class="flex items-center">
                              <div class="mr3">
                                <p class="lh-copy f5">This action is not reversible.</p>
                              </div>
                              <input class="bg-red white ba bw b--dark-red f5 b pv2 ph3 ma0 grow" type="submit" value="Delete account">
                            </div>
                          </div>
                        </div>
                      </div>
                    `
                  })

                  document.body.appendChild(dialogEl)
                },
                size: 'none'
              })}

              <div class="ml3">
                <a id="delete-account"></a>
                <p class="lh-copy f5 dark-gray">
                  Request your account to be deleted.
                </p>
              </div>
            </section>
          </div>
        </div>
      </section>
    </div>
  `
}
