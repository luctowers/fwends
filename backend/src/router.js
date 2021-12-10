const Router = require('@koa/router');

const router = new Router();

router.get('/api', (ctx) => {
  ctx.body = 'Hello from backend!';
})

router.use('/api/packs', require('./routes/packs').routes())

module.exports = router;
