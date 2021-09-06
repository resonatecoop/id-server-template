const html = require('choo/html')

module.exports = (state, emit) => {
  return html`
    <div class="flex flex-column">
      <h2 class="f3 fw1 mt3 near-black near-black--light light-gray--dark lh-title">404 not found</h2>

      <p>This page does not exist yet.</p>
    </div>
  `
}
