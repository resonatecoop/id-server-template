require('babel-polyfill')

const choo = require('choo')
const app = choo()
const html = require('choo/html')
const RandomArtistsGrid = require('./components/artists/random-grid')
const Login = require('./components/forms/login')

app.use(() => {})

app.route('*', (state, emit) => {
  const grid = state.cache(RandomArtistsGrid, 'random-artists').render()
  const login = state.cache(Login, 'login').render()

  return html`
    <div id="app">
      <main class="flex flex-row-reverse flex-auto relative vh-100">
        <div class="flex flex-column flex-auto w-100">
          ${grid}
          <div class="flex flex-column flex-auto items-center justify-center min-vh-100 mh3 pt6 pb7">
            <div class="bg-white black bg-black--dark white--dark bg-white--light black--light z-1 w-100 w-auto-l shadow-contour ph4 pt4 pb3">
              <div class="flex flex-column flex-auto">
                <svg viewBox="0 0 16 16" class="icon icon-logo icon--sm icon icon--lg fill-black fill-white--dark fill-black--light">
                  <use xlink:href="#icon-logo" />
                </svg>
                <h2 class="f3 fw1 mt2 near-black near-black--light light-gray--dark lh-title">Log In</h2>
                ${login}
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  `
})

module.exports = app.mount('#app')
