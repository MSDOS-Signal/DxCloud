// Web Terminal E2E：模拟浏览器 xterm.js 行为（onopen 立即发 resize + 输入）
import WebSocket from 'ws'

const base = 'http://localhost/api/v1'

async function main() {
  const loginRes = await fetch(base + '/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'admin', password: 'Admin@123456' }),
  })
  const adm = (await loginRes.json()).data

  const execRes = await fetch(base + '/ecs/8/exec', {
    method: 'POST',
    headers: { Authorization: 'Bearer ' + adm.access_token },
  })
  const { token } = (await execRes.json()).data
  console.log('token obtained, len =', token.length)

  const ws = new WebSocket('ws://localhost/ws/v1/ecs/8/terminal?token=' + token)
  let buf = ''
  let closed = false

  ws.on('open', () => {
    console.log('ws open')
    ws.send(JSON.stringify({ type: 'resize', cols: 140, rows: 40 }))
    ws.send(Buffer.from('id; stty size\r'))
    setTimeout(() => {
      console.log('--- received ---')
      console.log(JSON.stringify(buf))
      const okId = buf.includes('uid=')
      const okResize = buf.includes('40 140')
      console.log('id output:', okId, '| resize rows=40 cols=140:', okResize)
      ws.close()
      process.exit(okId && okResize ? 0 : 1)
    }, 5000)
  })
  ws.on('message', (d) => {
    buf += d.toString()
    if (buf.includes('uid=') && buf.includes('140')) {
      // 提前满足条件也等满，避免 close 竞态
    }
  })
  ws.on('close', (code, reason) => {
    console.log('ws closed', code, reason.toString())
    if (!closed) {
      closed = true
    }
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
