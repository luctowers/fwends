const Router = require('@koa/router');

const router = new Router();

router.get('/api', (ctx) => {
  ctx.body = 'Hello from backend!';
})

router.use('/api/pack', require('./routes/pack').routes())

module.exports = router;
