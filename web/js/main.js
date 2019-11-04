const choo = require('choo')
const app = choo()
const html = require('choo/html')

app.use((state, emitter) => {
  //
})

app.route('*', (state, emit) => {
  return html`
    <div id="app">

    </div>
  `
})

module.exports = app.mount('#app')
