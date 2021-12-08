const html = require('choo/html')
const format = require('date-fns/format')

module.exports = (state, emit) => {
  const total = state.shares.reduce((acc, share) => {
    return acc + share.amount
  }, 0)

  return html`
    <div class="flex flex-column">
      <h2 class="lh-title f3 fw1">Your memberships</h2>
      <p class="lh-copy f5 measure">
        You may cancel or renew an active membership at any time.<br>
        No refund will be be made in case of a cancelation.
      </p>
      <div>
        <div class="flex flex-column flex-auto pb6">
          <div>
            <div class="overflow-auto ph4 ba bw b--black-20 pv4">
              <table class="f6 w-100 mw8 center" cellspacing="0">
                <thead>
                  <tr>
                    <th class="fw6 bb b--black-20 tl pb3 pr3 bg-white">Name</th>
                    <th class="fw6 bb b--black-20 tl pb3 pr3 bg-white">Date From</th>
                    <th class="fw6 bb b--black-20 tl pb3 pr3 bg-white">Until</th>
                    <th class="fw6 bb b--black-20 tl pb3 pr3 bg-white">Contribution</th>
                    <th class="fw6 bb b--black-20 tl pb3 pr3 bg-white"></th>
                  </tr>
                </thead>
                <tbody class="lh-copy">
                  ${state.memberships.map(membership => html`
                    <tr>
                      <td class="pv3 pr3 bb b--black-20${!membership.active ? ' o-50' : ''}">
                        ${membership.name}
                      </td>
                      <td class="pv3 pr3 bb b--black-20${!membership.active ? ' o-50' : ''}">
                        ${format(new Date(membership.dateFrom), 'dd MMMM yyyy')}
                      </td>
                      <td class="pv3 pr3 bb b--black-20${!membership.active ? ' o-50' : ''}">
                        ${format(new Date(membership.dateTo), 'dd MMMM yyyy')}
                      </td>
                      <td class="pv3 pr3 bb b--black-20${!membership.active ? ' o-50' : ''}">
                        ${membership.contribution}
                      </td>
                      <td class="pv3 pr3 bb b--black-20">
                        ${membership.active
                          ? html`
                            <form action="" method="POST">
                              <input type="hidden" name="gorilla.csrf.Token" value=${state.csrfToken}>
                              <input type="hidden" name="_method" value="DELETE" />
                              <input type="hidden" name="id" value="${membership.subscriptionID}" />
                              <div class="flex flex-auto">
                                <button style="outline:solid 1px var(--near-black);outline-offset:-1px" type="submit" class="bg-white dib bn pa1 flex-shrink-0 f6 grow">
                                  Cancel
                                </button>
                              </div>
                            </form>
                          `
                          : ''}
                      </td>
                    </tr>
                  `)}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
      <h2 class="lh-title f3 fw1">Your shares</h2>
      <div>
        <div class="flex flex-column flex-auto pb6">
          <div>
            <div class="overflow-auto ph4 ba bw b--black-20 pv4">
              <table class="f6 w-100 mw8 center" cellspacing="0">
                <thead>
                  <tr>
                    <th></th>
                    <th class="fw1 bb b--black-20 tl pb3 pr3 f4">Amount (1€ par value)</th>
                    <th class="fw1 bb b--black-20 tl pb3 pr3 f4">Date Purchased</th>
                  </tr>
                </thead>
                <tbody class="lh-copy">
                  ${state.shares.map(share => html`
                    <tr>
                      <td></td>
                      <td class="pv3 pr3 bb b--black-20">
                        ${share.amount}
                      </td>
                      <td class="pv3 pr3 bb b--black-20">
                        ${format(new Date(share.datePurchased), 'dd MMMM yyyy')}
                      </td>
                    </tr>
                  `)}
                </tbody>
                <tfoot>
                  <tr>
                    <th scope="row">Total*</th>
                    <td>${total}</td>
                  </tr>
                </tfoot>
              </table>
            </div>
            <p class="lh-copy f5">* For shares bought before xxx, you may contact us to receive a full copy of the data.</p>
          </div>
        </div>
      </div>
    </div>
  `
}
