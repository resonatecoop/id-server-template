/* global fetch */

const choo = require('choo')
const app = choo({ href: false }) // disable choo href routing

const { isBrowser } = require('browser-or-node')
const setTitle = require('./src/lib/title')

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

// main app store
app.use((state, emitter) => {
  state.profile = state.profile || {
    displayName: ''
  }

  state.usergroup = {
    member: 'artist',
    listener: 'listener',
    fans: 'listener',
    'label-owner': 'label',
    admin: 'admin',
    uploader: 'uploader',
    volunteer: 'volunteer'
  }[state.profile.role]

  state.clients = state.clients || [
    {
      connectUrl: 'https://stream.resonate.coop/api/user/connect/resonate',
      name: 'Player',
      description: 'stream.resonate.coop'
    },
    {
      connectUrl: 'https://dash.resonate.coop/api/user/connect/resonate',
      name: 'Artist Dashboard',
      description: 'dash.resonate.coop'
    }
  ]

  emitter.on(state.events.DOMCONTENTLOADED, () => {
    emitter.emit(`route:${state.route}`)
    setMeta()
  })

  emitter.on('set:usergroup', (usergroup) => {
    state.usergroup = usergroup
    emitter.emit(state.events.RENDER)
  })

  emitter.on('route:profile', async () => {
    try {
      const response = await (await fetch(`https://${process.env.API_DOMAIN}/v2/user/profile`, {
        headers: {
          Authorization: 'Bearer ' + state.profile.token
        }
      })).json()

      state.profile = Object.assign({}, state.profile, response.data)

      emitter.emit(state.events.RENDER)
    } catch (err) {
      console.log(err)
    }
  })

  emitter.on(state.events.NAVIGATE, () => {
    emitter.emit(`route:${state.route}`)
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

app.use(require('./src/plugins/notifications')())

// layouts
const layout = require('./src/layouts/default')
const layoutNarrow = require('./src/layouts/narrow')

// choo routes
app.route('/', layout(require('./src/views/home')))
app.route('/authorize', layoutNarrow(require('./src/views/authorize')))
app.route('/join', layoutNarrow(require('./src/views/join')))
app.route('/login', layoutNarrow(require('./src/views/login')))
app.route('/password-reset', layoutNarrow(require('./src/views/password-reset')))
app.route('/email-confirmation', layoutNarrow(require('./src/views/email-confirmation')))
app.route('/account-settings', layout(require('./src/views/account-settings')))
app.route('/welcome', layoutNarrow(require('./src/views/welcome')))
app.route('/profile', layoutNarrow(require('./src/views/profile')))
app.route('/profile/new', layoutNarrow(require('./src/views/profile/new')))
app.route('*', layoutNarrow(require('./src/views/404')))

module.exports = app.mount('#app')
