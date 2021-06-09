const html = require('choo/html')

const Signup = require('../components/forms/signup')

module.exports = (state, emit) => {
  const signup = state.cache(Signup, 'signup')

  return html`
    <div class="flex flex-column">
      <h2 class="f3 fw1 mt3 near-black near-black--light light-gray--dark lh-title">Join now</h2>
      ${signup.render()}
      <p class="f6 lh-copy measure">
        By signing up, you accept the <a class="link b" href="https://resonate.is/terms-conditions/" target="_blank" rel="noopener">Terms and Conditions</a> and acknowledge the <a class="link b" href="https://resonate.is/privacy-policy/" target="_blank">Privacy Policy</a>.
      </p>
    </div>
  `
}
