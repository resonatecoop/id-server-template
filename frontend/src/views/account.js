/* global fetch */

const html = require('choo/html')

const icon = require('@resonate/icon-element')
// const Dialog = require('@resonate/dialog-component')
// const Button = require('@resonate/button-component')

const UpdateProfileForm = require('../components/forms/profile')
// const UpdatePasswordForm = require('../components/forms/passwordUpdate')

/**
 * Account settings
 * @param {Object} state Choo state
 * @param {Function} emit Emit choo event (nanobus)
 */
module.exports = (state, emit) => {
  // const deleteButton = new Button('delete-profile-button')

  return html`
    <div class="flex flex-column items-center justify-center w-100 mh3 mh0-ns">
      <section id="account-settings" class="flex flex-column">
        <div class="flex flex-column flex-auto pt4 ph3 mw6 ph0-l">
          ${!state.profile.complete ? icon('logo', { size: 'lg' }) : ''}
          <h2 class="lh-title f3 fw1">${state.profile.complete ? 'Update' : 'Create'} your account</h2>
          <div>
            <div class="flex flex-column flex-auto pb6">
              ${state.cache(UpdateProfileForm, 'update-account').render({
                data: state.profile || {}
              })}
            </div>
          </div>
        </div>
      </section>
    </div>
  `
}
