const html = require('choo/html')

module.exports = (view) => {
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
