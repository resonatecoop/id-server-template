const html = require('choo/html')

module.exports = (state, emit) => {
  return html`
    <div class="flex flex-column ph2 ph0-ns mw6 mt5 center pb6">
      <article>
        <h1 class="lh-title fw1 f2">Welcome to Resonate</h1>

        <p class="mb3">You’ve just made a private profile. </p>

        <p class="mb3">If you want to share music and playlists, you’ll need a public one, too. </p>

        <p class="mb3">Create as many public profiles as you’d like. You’ll always have control over what gets shared and what stays private.</p>

        <a href="/profile" style="outline:solid 1px var(--near-black);outline-offset:-1px" class="link dib b pv2 ph4 mb1">Create your profile</a>

        <p>No thanks, <a href="https://beta.stream.resonate.coop/api/v2/user/connect/resonate" class="link dib b">I want to listen to music privately</a>.</p>
      </article>
    </div>
  `
}
