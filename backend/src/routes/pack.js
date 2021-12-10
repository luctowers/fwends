const Router = require('@koa/router');
var AWS = require('aws-sdk');

var s3  = new AWS.S3({
  accessKeyId: process.env.OBJ_ACCESS_KEY,
  secretAccessKey: process.env.OBJ_SECRET_KEY,
  endpoint: process.env.OBJ_ENDPOINT,
  s3ForcePathStyle: true, // needed with minio?
  signatureVersion: 'v4'
});

const router = new Router();

module.exports = router;
