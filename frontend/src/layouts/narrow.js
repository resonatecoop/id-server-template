const html = require('choo/html')

module.exports = (view) => {
  return (state, emit) => {
    return html`
      <div id="app">
        <main class="flex flex-column flex-auto items-center justify-center min-vh-100 mh3 pb6">
          <div class="flex flex-column w-100 w-auto-l ph0 ph4-ns pt4 pb3">
            ${view(state, emit)}
          </div>
        </main>
      </div>
    `
  }
}
