const choo = require('choo')
const app = choo({ href: false })
const html = require('choo/html')
const Authorize = require('./components/forms/authorize')
const Login = require('./components/forms/login')
const Signup = require('./components/forms/signup')
const PasswordReset = require('./components/forms/passwordReset')
const PasswordResetUpdatePassword = require('./components/forms/passwordResetUpdatePassword')
const UpdateProfileForm = require('./components/forms/profile')
const UpdatePasswordForm = require('./components/forms/passwordUpdate')
const Notifications = require('./components/notifications')
const CountrySelect = require('./components/select-country-list')
const Dialog = require('@resonate/dialog-component')
const Button = require('@resonate/button-component')
const { isBrowser } = require('browser-or-node')
const setTitle = require('./lib/title')
const navigateToAnchor = (e) => {
  const el = document.getElementById(e.target.hash.substr(1))
  if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' })
  e.preventDefault()
}

if (isBrowser) {
  require('web-animations-js/web-animations.min')

  window.localStorage.DISABLE_NANOTIMING = process.env.DISABLE_NANOTIMING === 'yes'
  window.localStorage.logLevel = process.env.LOG_LEVEL

  if (process.env.NODE_ENV !== 'production') {
    app.use(require('choo-devtools')())
  }

  if ('Notification' in window) {
    app.use(require('choo-notification')())
  }
}

app.use(require('choo-meta')())

app.use((state, emitter) => {
  state.clients = state.clients || [
    {
      connectUrl: 'https://upload.resonate.is/api/user/connect/resonate',
      name: 'Upload Tool',
      description: 'for creators'
    }
  ]
  state.profile = state.profile || {
    displayName: ''
  }

  emitter.on(state.events.DOMCONTENTLOADED, () => {
    setMeta()
  })

  emitter.on(state.events.NAVIGATE, () => {
    setMeta()
  })

  function setMeta () {
    const title = {
      '*': 'Page not found',
      '/': 'Apps',
      login: 'Log In',
      authorize: 'Authorize',
      profile: 'Profile',
      'password-reset': 'Password reset',
      join: 'Join'
    }[state.route]

    if (!title) return

    state.shortTitle = title

    const fullTitle = setTitle(title)

    emitter.emit('meta', {
      title: fullTitle
    })
  }
})

app.use((state, emitter) => {
  state.messages = state.messages || []

  emitter.on(state.events.DOMCONTENTLOADED, _ => {
    emitter.on('notification:denied', () => {
      emitter.emit('notify', {
        type: 'warning',
        timeout: 6000,
        message: 'Notifications are blocked, you should modify your browser site settings'
      })
    })

    emitter.on('notification:granted', () => {
      emitter.emit('notify', {
        type: 'success',
        message: 'Notifications are enabled'
      })
    })

    emitter.on('notify', (props) => {
      const { message } = props

      if (!state.notification.permission) {
        const dialog = document.querySelector('dialog')
        const name = dialog ? 'notifications' : 'notifications-dialog'
        const notifications = state.cache(Notifications, name)
        const host = props.host || (dialog || document.body)

        if (notifications.element) {
          notifications.add(props)
        } else {
          const el = notifications.render({
            size: dialog ? 'small' : 'default'
          })
          host.insertBefore(el, host.firstChild)
          notifications.add(props)
        }
      } else {
        emitter.emit('notification:new', message)
      }
    })
  })
})

const layout = (view) => {
  return (state, emit) => {
    return html`
      <div id="app" class="flex flex-column pb6">
        <main class="flex flex-auto">
          ${view(state, emit)}
        </main>
      </div>
    `
  }
}

const layoutNarrow = (view) => {
  return (state, emit) => {
    return html`
      <div id="app">
        <main class="flex flex-auto relative">
          <div class="flex flex-column flex-auto w-100">
            <div class="flex flex-column flex-auto items-center justify-center min-vh-100 mh3 pt6 pb6">
              <div class="bg-white black bg-black--dark white--dark bg-white--light black--light z-1 w-100 w-auto-l ph4 pt4 pb3">
                <div class="flex flex-column flex-auto">
                  <svg viewBox="0 0 16 16" class="icon icon-logo icon--sm icon icon--lg fill-black fill-white--dark fill-black--light">
                    <use xlink:href="#icon-logo" />
                  </svg>
                  ${view(state, emit)}
                </div>
              </div>
            </div>
          </div>
        </main>
      </div>
    `
  }
}

app.route('/authorize', layoutNarrow((state, emit) => {
  const authorize = state.cache(Authorize, 'authorize')

  return html`
    <div class="flex flex-column">
      <h2 class="f3 fw1 mt3 near-black near-black--light light-gray--dark lh-title">Authorize</h2>
      ${authorize.render()}
    </div>
  `
}))

app.route('/join', layoutNarrow((state, emit) => {
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
}))

app.route('/login', layoutNarrow((state, emit) => {
  const login = state.cache(Login, 'login')

  return html`
    <div class="flex flex-column">
      <h2 class="f3 fw1 mt3 near-black near-black--light light-gray--dark lh-title">Log In</h2>
      ${login.render()}
    </div>
  `
}))

app.route('/', layout((state, emit) => {
  return html`
    <div class="flex flex-auto flex-column w-100 pb6">
      <article class="mh2 mt3 cf">
        ${state.clients.map(({ connectUrl, name, description }) => {
          return html`
            <div class="fl w-50 pa2 mw4-ns mw5-l">
              <a href=${connectUrl} class="link db aspect-ratio aspect-ratio--1x1 dim ba bw b--mid-gray">
                <div class="flex flex-column justify-center aspect-ratio--object pa2 pa3-ns pa4-l">
                  <span class="f3 f4-ns f3-l lh-title">${name}</span>
                  <span class="f4 f5-ns f4-l lh-copy">${description}</span>
                </div>
              </a>
            </div>
          `
        })}
        <div class="fl w-50 pa2 mw4-ns mw5-l">
          <a href="/apps" class="link db aspect-ratio aspect-ratio--1x1 dim bg-gray ba bw b--mid-gray black">
            <div class="flex flex-column justify-center aspect-ratio--object pa2 pa3-ns pa4-l">
              <span class="f4 f5-ns f4-l lh-copy">Register a new app</span>
            </div>
          </a>
        </div>
      </article>

      <p class="ml3 lh-copy measure f4 f5-ns f4-l">Not a member yet? <a class="link b" href="/join">Join now!</a></p>
    </div>
  `
}))

app.route('/apps', layoutNarrow((state, emit) => {
  return html`
    <div class="flex flex-column">
      <h2 class="f3 fw1 mt3 near-black near-black--light light-gray--dark lh-title">Register a new app</h2>
    </div>
  `
}))

app.route('/password-reset', layoutNarrow((state, emit) => {
  const passwordReset = state.cache(PasswordReset, 'password-reset')
  const passwordResetUpdatePassword = state.cache(PasswordResetUpdatePassword, 'password-reset-update')

  return html`
    <div class="flex flex-column">
      <h2 class="f3 fw1 mt3 near-black near-black--light light-gray--dark lh-title">Reset your password</h2>

      ${state.query.token ? passwordResetUpdatePassword.render({
        token: state.query.token
      }) : passwordReset.render()}
    </div>
  `
}))

/**
 * Note: keep this as placeholder ?
 */

app.route('/email-confirmation', layoutNarrow((state, emit) => {
  return html`
    <div class="flex flex-column">
    </div>
  `
}))

app.route('/profile', layout((state, emit) => {
  const deleteButton = new Button('delete-profile-button')

  return html`
    <div class="flex flex-column w-100 mh3 mh0-ns">
      <section id="account-settings" class="flex flex-column">
        <h2 class="lh-title pl3 f2 fw1">Account settings</h2>
        <div class="flex flex-column flex-row-l">
          <div class="w-50 w-third-l ph3">
            <nav class="sticky z-1 flex flex-column" style="top:3rem">
              <ul class="list ma0 pa0 flex flex-column">
                <li class="mb2">
                  <a class="link" href="#account-info" onclick=${navigateToAnchor}>Account</a>
                </li>
                <li class="mb2">
                  <a class="link" href="#change-country" onclick=${navigateToAnchor}>Location</a>
                </li>
                <li class="mb2">
                  <a class="link" href="#change-password" onclick=${navigateToAnchor}>Change password</a>
                </li>
                <li>
                  <a class="link" href="#delete-account" onclick=${navigateToAnchor}>Delete account</a>
                </li>
              </ul>
            </nav>
          </div>
          <div class="flex flex-column flex-auto ph3 pt4 pt0-l mw6 ph0-l">
            <div class="ph3">
              <a id="account-info" class="absolute" style="top:-120px"></a>
              ${state.cache(UpdateProfileForm, 'update-profile').render({
                data: state.profile || {}
              })}
            </div>

            <div class="ph3">
              <h3 class="f3 fw1 lh-title relative mb3">
                Location
                <a id="change-country" class="absolute" style="top:-120px"></a>
              </h3>
              ${state.cache(CountrySelect, 'update-country').render({
                country: state.profile.country || ''
              })}
            </div>

            <div class="ph3">
              <h3 class="f3 fw1 lh-title relative mb3">
                Change password
                <a id="change-password" class="absolute" style="top:-120px"></a>
              </h3>
              ${state.cache(UpdatePasswordForm, 'update-password-form').render()}
            </div>

            <div class="flex w-100 items-center ph3">
              ${deleteButton.render({
                type: 'button',
                prefix: 'bg-white ba bw b--dark-gray f5 b pv3 ph3 w-100 mw5 grow',
                text: 'Delete account',
                style: 'none',
                onClick: () => {
                  const dialog = state.cache(Dialog, 'delete-account-dialog')
                  const dialogEl = dialog.render({
                    title: 'Delete account',
                    prefix: 'dialog-default dialog--sm',
                    onClose: async (e) => {
                      if (e.target.returnValue === 'Delete account') {
                        try {
                          // do something
                        } catch (err) {
                          emit('error', err)
                        }
                      }

                      dialog.destroy()
                    },
                    content: html`
                      <div class="flex flex-column">
                        <p class="lh-copy f5 b">Are you sure you want to delete your Resonate account ?</p>

                        <div class="flex">
                          <div class="flex items-center">
                            <input class="bg-white black ba bw b--near-black f5 b pv2 ph3 ma0 grow" type="submit" value="Not really">
                          </div>
                          <div class="flex flex-auto w-100 justify-end">
                            <div class="flex items-center">
                              <div class="mr3">
                                <p class="lh-copy f5">This action is not reversible.</p>
                              </div>
                              <input class="bg-red white ba bw b--dark-red f5 b pv2 ph3 ma0 grow" type="submit" value="Delete account">
                            </div>
                          </div>
                        </div>
                      </div>
                    `
                  })

                  document.body.appendChild(dialogEl)
                },
                size: 'none'
              })}

              <div class="ml3">
                <a id="delete-account"></a>
                <p class="lh-copy f5 dark-gray">
                  This will delete your account and all associated profiles.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>
  `
}))

app.route('*', layout((state, emit) => {
  return html`
    <div>
      <h2>404</h2>
    </div>
  `
}))

module.exports = app.mount('#app')
