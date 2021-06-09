const html = require('choo/html')

module.exports = (view) => {
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
