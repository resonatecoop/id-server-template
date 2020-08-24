const choo = require('choo')
const app = choo()
const html = require('choo/html')
const RandomArtistsGrid = require('./components/artists/random-grid')
const Authorize = require('./components/forms/authorize')
const Login = require('./components/forms/login')
const Footer = require('./components/footer')
const Signup = require('./components/forms/signup')
const setTitle = require('./lib/title')

if (process.env.NODE_ENV !== 'production') {
  app.use(require('choo-devtools')())
}

app.use(require('choo-meta')())

app.use((state, emitter) => {
  emitter.on(state.events.NAVIGATE, () => {
    setMeta()
  })

  function setMeta () {
    const title = {
      '*': 'Login',
      authorize: 'Authorize',
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

const layout = (view) => {
  return (state, emit) => {
    return html`
      <div id="app" class="flex flex-column">
        <main class="flex flex-auto">
          ${view(state, emit)}
        </main>
        ${state.cache(Footer, 'footer').render()}
      </div>
    `
  }
}

const layoutWithGrid = (view) => {
  return (state, emit) => {
    const grid = state.cache(RandomArtistsGrid, 'random-artists').render()

    return html`
      <div id="app">
        <main class="flex flex-auto relative">
          <div class="flex flex-column flex-auto w-100">
            ${grid}
            <div class="flex flex-column flex-auto items-center justify-center min-vh-100 mh3 pv6">
              <div class="bg-white black bg-black--dark white--dark bg-white--light black--light z-1 w-100 w-auto-l shadow-contour ph4 pt4 pb3">
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

app.route('/authorize', layoutWithGrid((state, emit) => {
  const authorize = state.cache(Authorize, 'authorize')

  return html`
    <div class="flex flex-column">
      <h2 class="f3 fw1 mt2 near-black near-black--light light-gray--dark lh-title">Authorize</h2>
      ${authorize.render()}
    </div>
  `
}))

app.route('/join', layoutWithGrid((state, emit) => {
  const signup = state.cache(Signup, 'signup')

  return html`
    <div class="flex flex-column">
      <h2 class="f3 fw1 mt2 near-black near-black--light light-gray--dark lh-title">Join now</h2>
      ${signup.render()}
    </div>
  `
}))

app.route('/login', layoutWithGrid((state, emit) => {
  const login = state.cache(Login, 'login')

  return html`
    <div class="flex flex-column">
      <h2 class="f3 fw1 mt2 near-black near-black--light light-gray--dark lh-title">Log In</h2>
      ${login.render()}
    </div>
  `
}))

app.route('/', layout((state, emit) => {
  const services = [
    {
      hostname: 'stream.resonate.localhost',
      pathname: '/api/user/connect/resonate',
      name: 'Player',
      description: 'fair streaming'
    },
    {
      hostname: 'upload.resonate.localhost',
      pathname: '/api/user/connect/resonate',
      name: 'Upload Tool',
      description: 'for creators'
    }
  ]

  return html`
    <div class="flex flex-auto flex-column w-100 pb6">
      <article class="mh2 cf">
        <h1 class="ml2 f2 f1-l lh-title fw3">Play fair</h1>
        <h2 class="ml2 f4 f3-l fw4">The community-owned music network.</h2>

        ${services.map(({ pathname, hostname, name, description }) => {
          const url = new URL(pathname, `https://${hostname}`)

          return html`
            <div class="fl w-50 pa2 mw4-ns mw5-l">
              <a href=${url.href} class="link db aspect-ratio aspect-ratio--1x1 dim ba bw b--near-black">
                <div class="flex flex-column justify-center aspect-ratio--object pa2 pa3-ns pa4-l">
                  <span class="f3 f4-ns f3-l lh-title">${name}</span>
                  <span class="f4 f5-ns f4-l lh-copy">${description}</span>
                </div>
              </a>
            </div>
          `
        })}
        <div class="fl w-50 pa2 mw4-ns mw5-l">
          <a href="/apps/create" class="link db aspect-ratio aspect-ratio--1x1 dim bg-gray black">
            <div class="flex flex-column justify-center aspect-ratio--object pa2 pa3-ns pa4-l">
              <span class="f4 f5-ns f4-l lh-copy">Build something new</span>
            </div>
          </a>
        </div>
      </article>

      <p class="ml3 lh-copy measure f4 f5-ns f4-l">Not a member yet? <a class="link b" href="/join">Join now!</a></p>
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
