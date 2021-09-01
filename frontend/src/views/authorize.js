const html = require('choo/html')

const Authorize = require('../components/forms/authorize')

module.exports = (state, emit) => {
  const authorize = state.cache(Authorize, 'authorize')

  return html`
    <div class="flex flex-column">
      <h2 class="f3 fw1 mt3 near-black near-black--light light-gray--dark lh-title">Continue to ${state.applicationName}</h2>
      ${authorize.render()}
    </div>
  `
}
