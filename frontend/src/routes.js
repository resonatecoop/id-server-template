const layout = require('./layouts/default')
const layoutNarrow = require('./layouts/narrow')

/**
 * @description Choo routes
 * @param {Object} app Choo app
 */
function routes (app) {
  app.route('/', layout(require('./views/home')))
  app.route('/authorize', layoutNarrow(require('./views/authorize')))
  app.route('/join', layoutNarrow(require('./views/join')))
  app.route('/login', layoutNarrow(require('./views/login')))
  app.route('/password-reset', layoutNarrow(require('./views/password-reset')))
  app.route('/email-confirmation', layoutNarrow(require('./views/email-confirmation')))
  app.route('/account-settings', layout(require('./views/account-settings')))
  app.route('/welcome', layoutNarrow(require('./views/welcome')))
  app.route('/profile', layoutNarrow(require('./views/profile')))
  app.route('/profile/new', layoutNarrow(require('./views/profile/new')))
  app.route('*', layoutNarrow(require('./views/404')))
}

module.exports = routes
