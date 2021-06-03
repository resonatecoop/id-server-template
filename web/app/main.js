const choo = require('choo')
const app = choo({ href: false }) // disable choo href routing

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
    setMeta()
  })

  emitter.on('set:usergroup', (usergroup) => {
    state.usergroup = usergroup
    emitter.emit(state.events.RENDER)
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

app.use(require('./plugins/notifications')())

// layouts
const layout = require('./layouts/default')
const layoutNarrow = require('./layouts/narrow')

// choo routes
app.route('/', layout(require('./views/home')))
app.route('/authorize', layoutNarrow(require('./views/authorize')))
app.route('/join', layoutNarrow(require('./views/join')))
app.route('/login', layoutNarrow(require('./views/login')))
app.route('/password-reset', layoutNarrow(require('./views/password-reset')))
app.route('/email-confirmation', layoutNarrow(require('./views/email-confirmation')))
app.route('/account-settings', layout(require('./views/account-settings')))
app.route('/welcome', layoutNarrow(require('./views/welcome')))
app.route('/profile', layoutNarrow(require('./views/profile')))
app.route('/profile/new', layoutNarrow(require('./views/profile/new')))
app.route('*', layoutNarrow(require('./views/404')))

module.exports = app.mount('#app')
