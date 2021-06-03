const html = require('choo/html')
const ProfileTypeForm = require('../../components/forms/profile-type')

/**
 * Render view to create new/additional profile
 * @param {Object} state Choo state
 * @param {Function} emit Emit choo event (nanobus)
 */
module.exports = (state, emit) => {
  return html`
    <div class="flex flex-column ph2 ph0-ns mw6 mt5 center pb6">
      ${state.cache(ProfileTypeForm, 'new-profile-type-form').render({ value: state.usergroup })}
    </div>
  `
}
