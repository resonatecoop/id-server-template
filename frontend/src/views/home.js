const html = require('choo/html')

module.exports = (state, emit) => {
  return html`
    <div class="flex flex-auto flex-column w-100 pb6">
      <article class="mh2 mt3 cf">
        ${state.clients.map(({ connectUrl, name, description }) => {
          return html`
            <div class="fl w-100 w-50-ns w-33-l pa2">
              <a href=${connectUrl} class="link db aspect-ratio aspect-ratio--1x1 dim ba bw b--near-black">
                <dl class="flex flex-column justify-center aspect-ratio--object pa3 pa4-l">
                  <dt class="f3 lh-title">${name}</dt>
                  <dd class="ma0 f4 f5-ns f4-l lh-copy">${description}</dd>
                </dl>
              </a>
            </div>
          `
        })}
      </article>

      <p class="ml3 lh-copy measure f4 f5-ns f4-l">Not a member yet? <a class="link b" href="/join">Join now!</a></p>
    </div>
  `
}
