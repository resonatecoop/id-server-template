const html = require('choo/html')

module.exports = (state, emit) => {
  return html`
    <div class="flex flex-auto flex-column w-100 pb6">
      <article class="mh2 mt3 cf">
        ${state.clients.map(({ connectUrl, name, description }) => {
          return html`
            <div class="fl w-50 w-33-l pa2">
              <a href=${connectUrl} class="link db aspect-ratio aspect-ratio--1x1 dim ba bw b--mid-gray">
                <div class="flex flex-column justify-center aspect-ratio--object pa2 pa3-ns pa4-l">
                  <span class="f3 lh-title">${name}</span>
                  <span class="f4 lh-copy">${description}</span>
                </div>
              </a>
            </div>
          `
        })}
      </article>

      <p class="ml3 lh-copy measure f4 f5-ns f4-l">Not a member yet? <a class="link b" href="/join">Join now!</a></p>
    </div>
  `
}
