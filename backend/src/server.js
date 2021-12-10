const Koa = require('koa')
const router = require('./router')

const app = new Koa()
const port = 8080

app
  .use(router.routes())
  .use(router.allowedMethods());

app.listen(port, () => { console.log(`listening on port ${port}`) })
