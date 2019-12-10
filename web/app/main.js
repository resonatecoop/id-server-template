const choo = require('choo')
const app = choo()
const html = require('choo/html')
const RandomArtistsGrid = require('./components/artists/random-grid')
const Authorize = require('./components/forms/authorize')
const Login = require('./components/forms/login')

if (process.env.NODE_ENV !== 'production') {
  app.use(require('choo-devtools')())
}

const layout = (view) => {
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

app.route('/authorize', layout((state, emit) => {
  const authorize = state.cache(Authorize, 'authorize')

  return html`
    <div class="flex flex-column">
      <h2 class="f3 fw1 mt2 near-black near-black--light light-gray--dark lh-title">Authorize</h2>
      ${authorize.render()}
    </div>
  `
}))

app.route('*', layout((state, emit) => {
  const login = state.cache(Login, 'login')

  return html`
    <div class="flex flex-column">
      <h2 class="f3 fw1 mt2 near-black near-black--light light-gray--dark lh-title">Log In</h2>
      ${login.render()}
    </div>
  `
}))

module.exports = app.mount('#app')
