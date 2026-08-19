// Web Terminal E2E（参数化）：ECS_ID + ADMIN_TOKEN 由环境变量传入
import WebSocket from 'ws'

const base = 'http://localhost/api/v1'
const ecsId = process.env.ECS_ID
const token = process.env.ADMIN_TOKEN

if (!ecsId || !token) {
  console.error('usage: ECS_ID=<id> ADMIN_TOKEN=<jwt> node terminal-any.mjs')
  process.exit(2)
}

async function main() {
  const execRes = await fetch(`${base}/ecs/${ecsId}/exec`, {
    method: 'POST',
    headers: { Authorization: 'Bearer ' + token },
  })
  if (!execRes.ok) {
    console.error('exec issue failed:', execRes.status, await execRes.text())
    process.exit(2)
  }
  const { token: wsToken } = (await execRes.json()).data
  console.log('console token obtained, len =', wsToken.length)

  const ws = new WebSocket(`ws://localhost/ws/v1/ecs/${ecsId}/terminal?token=${wsToken}`)
  let buf = ''
  ws.on('open', () => {
    console.log('ws open')
    ws.send(JSON.stringify({ type: 'resize', cols: 140, rows: 40 }))
    ws.send(Buffer.from('id; stty size\r'))
    setTimeout(() => {
      const okId = buf.includes('uid=')
      const okResize = buf.includes('40 140')
      console.log('id output:', okId, '| resize rows=40 cols=140:', okResize)
      ws.close()
      process.exit(okId && okResize ? 0 : 1)
    }, 5000)
  })
  ws.on('message', (d) => {
    buf += d.toString()
  })
  ws.on('error', (e) => {
    console.log('ws error:', e.message)
    process.exit(2)
  })
}

main().catch((e) => {
  console.error('test error:', e)
  process.exit(2)
})
