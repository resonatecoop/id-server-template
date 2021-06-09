const html = require('choo/html')
const ArtistInfoForm = require('../../components/forms/artist-info')
const LabelInfoForm = require('../../components/forms/label-info')
const BasicInfoForm = require('../../components/forms/basic-info')
const ProfileTypeForm = require('../../components/forms/profile-type')

/**
 * Render view for artist, label and other profile forms
 * @param {Object} state Choo state
 * @param {Function} emit Emit choo event (nanobus)
 */
module.exports = (state, emit) => {
  const form = {
    artist: state.cache(ArtistInfoForm, 'artist-info'),
    label: state.cache(LabelInfoForm, 'label-info'),
    listener: state.cache(BasicInfoForm, 'basic-info')
  }[state.usergroup] || state.cache(ProfileTypeForm, 'profile-type-form')

  return html`
    <div class="flex flex-column ph2 ph0-ns mw6 mt5 center pb6">
      ${form.render({ value: state.usergroup })}
    </div>
  `
}
