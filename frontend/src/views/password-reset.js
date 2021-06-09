const html = require('choo/html')

const PasswordReset = require('../components/forms/passwordReset')
const PasswordResetUpdatePassword = require('../components/forms/passwordResetUpdatePassword')

module.exports = (state, emit) => {
  const passwordReset = state.cache(PasswordReset, 'password-reset')
  const passwordResetUpdatePassword = state.cache(PasswordResetUpdatePassword, 'password-reset-update')

  return html`
    <div class="flex flex-column">
      <h2 class="f3 fw1 mt3 near-black near-black--light light-gray--dark lh-title">Reset your password</h2>

      ${state.query.token
        ? passwordResetUpdatePassword.render({
          token: state.query.token
        })
        : passwordReset.render()
      }
    </div>
  `
}
