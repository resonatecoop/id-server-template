const html = require('choo/html')
const ProfileForm = require('../../components/forms/basic-info')

/**
 * Render view for artist, label and other profile forms
 * @param {Object} state Choo state
 * @param {Function} emit Emit choo event (nanobus)
 */
module.exports = (state, emit) => {
  return html`
    <div class="flex flex-column mw6 mt5 center pb6">
      ${state.cache(ProfileForm, 'profile-form').render({
        role: state.profile.role,
        profile: state.profile || {}
      })}
    </div>
  `
}
