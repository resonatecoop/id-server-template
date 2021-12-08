const html = require('choo/html')
const ProfileForm = require('../components/forms/basic-info')

/**
 * Render view for artist, label and other profile forms
 * @param {Object} state Choo state
 * @param {Function} emit Emit choo event (nanobus)
 */
module.exports = (state, emit) => {
  return html`
    <div class="flex flex-column flex-auto min-vh-100">
      ${state.cache(ProfileForm, 'profile-form').render({
        profile: state.profile || {}
      })}
    </div>
  `
}
