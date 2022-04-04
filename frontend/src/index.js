const choo = require('choo')
const nanochoo = require('nanochoo')
const initialState = window.initialState
  ? Object.assign({}, window.initialState)
  : {}
const app = choo({ href: false }) // disable choo href routing
window.initialState = initialState // hack to bring back initial state (should be deleted again by nanochoo)

const { isBrowser } = require('browser-or-node')
const setTitle = require('./lib/title')
const { getAPIServiceClientWithAuth } = require('@resonate/api-service')({
  apiHost: process.env.APP_HOST,
  base: process.env.API_BASE || '/api/v3'
})

const SearchOuter = require('./components/header')
const UserMenu = require('./components/user-menu')

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
    displayName: '',
    member: false
  }

  state.profile.avatar = state.profile.avatar || {}

  state.clients = state.clients || [
    {
      connectUrl: 'https://stream.resonate.coop/api/user/connect/resonate',
      name: 'Player',
      description: 'stream.resonate.coop'
    },
    {
      connectUrl: 'https://dash.resonate.coop/api/user/connect/resonate',
      name: 'Dashboard',
      description: 'dash.resonate.coop'
    }
  ]

  emitter.on(state.events.DOMCONTENTLOADED, () => {
    emitter.emit(`route:${state.route}`)
    setMeta()
  })

  emitter.on('route:account', () => {
    getUserProfile()
  })

  emitter.on('route:profile', () => {
    getUserProfile()
  })

  emitter.on(state.events.NAVIGATE, () => {
    emitter.emit(`route:${state.route}`)
    setMeta()
  })

  async function getUserProfile () {
    try {
      // get v2 api profile for legacy values (old nickname, avatar)
      const getClient = getAPIServiceClientWithAuth(state.token)
      const client = await getClient('profile')
      const result = await client.getUserProfile()

      const { body: response } = result
      const { data: userData } = response

      state.profile.nickname = userData.nickname
      state.profile.avatar = userData.avatar || {}

      emitter.emit(state.events.RENDER)
    } catch (err) {
      console.log(err.message)
      console.log(err)
    }
  }

  function setMeta () {
    const title = {
      '*': 'Page not found',
      '/': 'Apps',
      login: 'Log In',
      authorize: 'Authorize',
      profile: 'Create your profile',
      account: 'Update your account',
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

require('./routes')(app)

/*
 * Append search component to header (outside of main choo app)
 */
async function searchApp (initialState) {
  window.initialState = initialState

  const search = nanochoo()

  search.use((state, emitter, app) => {
    state.search = state.search || {
      q: ''
    }

    state.user = {
      token: state.token
    }
    state.params = {} // nanochoo does not have a router

    emitter.on('search', (q) => {
      const bang = q.startsWith('#')
      const pathname = bang ? '/tag' : '/search'
      const url = new URL(pathname, process.env.APP_HOST || 'http://localhost')
      const params = bang ? { term: q.split('#')[1] } : { q }
      url.search = new URLSearchParams(params)
      return window.open(url.href, '_blank')
    })
  })

  search.view((state, emit) => {
    // component id needs to be header to work correctly
    return state.cache(SearchOuter, 'header').render()
  })

  search.mount('#search-host')

  return Promise.resolve()
}

/*
 * Append usermenu app
 */
async function userMenuApp (initialState) {
  if (!document.getElementById('usermenu')) return

  window.initialState = initialState

  const usermenu = nanochoo()

  usermenu.use((state, emitter, app) => {
    state.params = {} // nanochoo does not have a router
  })

  usermenu.view((state, emit) => {
    return state.cache(UserMenu, 'usermenu').render({
      displayName: state.profile.displayName
    })
  })

  usermenu.mount('#usermenu')

  return Promise.resolve()
}

searchApp(initialState).then(() => {
  console.log('Loaded search app')
})

userMenuApp(initialState).then(() => {
  console.log('Loaded user menu')
})

module.exports = app.mount('#app')
