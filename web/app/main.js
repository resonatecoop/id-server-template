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
const AppForm = require('./components/forms/app')
const Notifications = require('./components/notifications')
const imagePlaceholder = require('./lib/image-placeholder')
const Dialog = require('@resonate/dialog-component')
const Button = require('@resonate/button-component')
const { isBrowser } = require('browser-or-node')
const setTitle = require('./lib/title')

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
  state.apps = state.apps || []
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

app.route('/apps', layout((state, emit) => {
  return html`
    <div class="flex flex-column flex-auto w-100">
      <div class="flex flex-column flex-auto items-center justify-center min-vh-100 mh3 pt6 pb6">
        <div class="bg-white black bg-black--dark white--dark bg-white--light black--light z-1 w-100 ph4 pt4 pb3">
          <div class="flex flex-column flex-row-l flex-auto">
            <div class="flex flex-auto flex-column ph3">
              <h2 class="lh-title f3 fw1">Your apps</h2>

              ${state.apps.map(app => {
                const { ID, key, applicationName: name, applicationUrl: url, applicationHostname: hostname, redirectUri } = app

                return html`
                  <div class="flex flex-column ba bw b--dark-gray mb4 ph3 pt4">
                    <fieldset class="ma0 pa0 ph3 bn">
                      <legend class="lh-copy mt3 f4 fw1 mb3">${name}</legend>
                      <div class="flex flex-column">
                        <label class="db" for="apps[${key}][ID]">Client ID</label>
                        <div class="mb3 flex">
                          <input name="apps[${key}][ID]" readonly disabled value=${ID} class="bg-black white bg-white--dark black--dark bg-black--light white--light placeholder--dark-gray input-reset w-100 bn pa3 valid">
                        </div>
                        <label class="db" for="apps[${key}][key]">Client Key</label>
                        <div class="mb3 flex">
                          <input name="apps[${key}][key]" readonly disabled value=${key} class="bg-black white bg-white--dark black--dark bg-black--light white--light placeholder--dark-gray input-reset w-100 bn pa3 valid">
                        </div>
                        <label class="db" for="apps[${key}][name]">Application Name</label>
                        <div class="mb3 flex">
                          <input name="apps[${key}][name]" readonly disabled value=${name} class="bg-black white bg-white--dark black--dark bg-black--light white--light placeholder--dark-gray input-reset w-100 bn pa3 valid">
                        </div>
                        <label class="db" for="apps[${key}][redirect_uri]">Redirect URI</label>
                        <div class="mb3 flex">
                          <input name="apps[${key}][redirect_uri]" readonly disabled value=${redirectUri} class="bg-black white bg-white--dark black--dark bg-black--light white--light placeholder--dark-gray input-reset w-100 bn pa3 valid">
                        </div>
                        <label class="db" for="apps[${key}][url]">Application URL</label>
                        <div class="mb3 flex">
                          <input name="apps[${key}][url]" readonly disabled value=${url} class="bg-black white bg-white--dark black--dark bg-black--light white--light placeholder--dark-gray input-reset w-100 bn pa3 valid">
                        </div>
                        <label class="db" for="apps[${key}][hostname]">Application Hostname</label>
                        <div class="mb3 flex">
                          <input name="apps[${key}][hostname]" readonly disabled value=${hostname} class="bg-black white bg-white--dark black--dark bg-black--light white--light placeholder--dark-gray input-reset w-100 bn pa3 valid">
                        </div>
                        <div class="mb3 flex justify-end">
                          <a href="/apps/${key}" class="link underline">Delete this app</a>
                        </div>
                      </div>
                    </fieldset>
                  </div>
                `
              })}
            </div>
            <div class="flex flex-auto flex-column ph3">
              <div class="sticky top-0">
                <h2 class="lh-title f3 fw1">Register a new app</h2>
                ${state.cache(AppForm, 'app-form').render()}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  `
}))

app.route('/apps/:id', layout((state, emit) => {
  return html`
    <div class="flex flex-column">
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
  const user = state.profile
  const src = imagePlaceholder(400, 400)
  const deleteButton = new Button('delete-profile-button')

  return html`
    <div class="flex flex-column w-100 mh3 mh0-ns">
      <section id="profile" class="flex flex-column flex-row-l">
        <div class="fl w-50 w-third-l pa3 mb4">
          <div class="sticky aspect-ratio aspect-ratio--1x1 bg-dark-gray bg-dark-gray--dark" style="top:3rem">
            <figure class="ma0">
              <img src=${src} width=400 height=400 class="aspect-ratio--object z-1" />
              <figcaption class="absolute bottom-0 truncate w-100 h2" style="top:100%;">
                ${user.displayName}
              </figcaption>
            </figure>
          </div>
        </div>

        <div class="flex flex-column flex-auto ph3 pt4 mw6 ph0-l">
          <div class="ph3">
            ${state.cache(UpdateProfileForm, 'update-profile-form').render({
              data: state.profile || {}
            })}
          </div>

          <div class="ph3">
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
                        await state.api.profile.remove()

                        window.location = `https://${process.env.APP_DOMAIN}/api/user/logout`
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
              <p class="lh-copy f5 dark-gray">
                This will delete your account and all associated profiles.
              </p>
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
