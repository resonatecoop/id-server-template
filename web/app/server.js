const jalla = require('jalla')
const path = require('path')
const app = jalla(path.join(__dirname, './main.js'))

const PORT = process.env.APP_PORT || 8081

app.listen(PORT)
