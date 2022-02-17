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
  app.route('/membership', layoutNarrow(require('./views/membership')))
  app.route('/checkout', layoutNarrow(require('./views/checkout')))
  app.route('/checkout/success', layoutNarrow(require('./views/checkout')))
  app.route('/checkout/cancel', layoutNarrow(require('./views/checkout')))
  app.route('/profile', layout(require('./views/profile')))
  app.route('/account', layout(require('./views/account')))
  app.route('/account-settings', layout(require('./views/account-settings')))
  app.route('*', layoutNarrow(require('./views/404')))
}

module.exports = routes
