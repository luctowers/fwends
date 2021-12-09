const express = require('express')
const app = express()
const port = 8080

app.get('/api', (req, res) => {
  res.send('Hello from backend!')
})

app.listen(port, () => { console.log(`listening on port ${port}`) })
