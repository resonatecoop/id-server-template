const html = require('choo/html')
const Login = require('../components/forms/login')

module.exports = (state, emit) => {
  const login = state.cache(Login, 'login')

  return html`
    <div class="flex flex-column">
      <h2 class="f3 fw1 mt3 near-black near-black--light light-gray--dark lh-title">Log In</h2>
      ${login.render()}
    </div>
  `
}
