const icon = require('@resonate/icon-element')
const html = require('choo/html')

module.exports = (state, emit) => {
  return html`
    <section id="checkout" class="flex flex-column">
      <div class="flex flex-column flex-auto pt4 ph3 mw6 ph0-l">
        <div class="flex">
          <a href="../web/account" class="link db flex items-center pv2 ph3 mb2">
            ${icon('arrow', { size: 'sm' })}
            ${!state.profile.complete
              ? html`
                <span class="pl2">
                  ${icon('logo', { size: 'lg' })}
                </span>`
              : ''}
          </a>
        </div>
        ${state.products.map(product => html`
          <article class="flex ba bw b--mid-gray pv3 ph4 mb3">
            <div>
              <figure class="ma0 w4 h4">
                <img src="${product.images[0]}">
                <figcaption class="clip">Product image</figcaption>
              </figure>
            </div>
            <div class="ph3">
              <p class="ma0 f3 fw1 lh-title">${product.name}</p>
              <dl>
                <dt class="clip">Description</dt>
                <dd class="ma0">
                  <p class="lh-copy f5">${product.description}</p>
                </dd>
                ${product.quantity > 0
                  ? html`
                    <dt class="dib mr2">Qty</dt>
                    <dd class="ma0 dib">
                      <p class="lh-copy f5 b">${product.quantity}</p>
                    </dd>
                  `
                  : ''}
              </dl>
            </div>
          </article>
        `)}
        <div class="flex items-center">
          <form action="" method="POST">
            <input type="hidden" name="gorilla.csrf.Token" value=${state.csrfToken}>
            <button type="submit" style="outline:solid 1px var(--near-black);outline-offset:-1px" type="submit" class="bg-white b dib bn pv3 ph5 flex-shrink-0 f5 grow">Checkout</button>
          </form>
          <p class="lh-copy pl3 f5">Powered by <a href="https://stripe.com" target="_blank" re="noreferer noopener" class="link b">Stripe</a></p>
        </div>
      </div>
    </div>
  `
}
