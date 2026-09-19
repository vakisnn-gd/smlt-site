const MOCK_PLAYERS = [
  {country: 'RS', name: 'SpaceRS', points: 1713.80, demon: 'Nullscapes', globalRank: 139, gdlId: 4804},
  {country: 'RU', name: 'Florned', points: 1643.07, demon: 'Defeated Circles', globalRank: 146, gdlId: 24272},
  {country: 'RU', name: 'Flik', points: 866.17, demon: 'Firework', globalRank: 229, gdlId: 6607},
  {country: 'RU', name: 'npoctou_gamer', points: 287.81, demon: 'Sevvend Clubstep', globalRank: 522, gdlId: 2383},
  {country: 'KZ', name: 'euphoria', points: 235.95, demon: 'poocubed', globalRank: 565},
  {country: 'RU', name: 'Tikys', points: 186.87, demon: 'Sevvend Clubstep', globalRank: 599},
  {country: 'OTHER', name: 'CandyCloud22', points: 120.70, demon: 'Cognition', globalRank: 716},
  {country: 'RU', name: 'toxik blaze', points: 102.05, demon: 'Tartarus', globalRank: 766},
  {country: 'RU', name: 'EfzEnn', points: 92.21, demon: 'Oblivion', globalRank: 804},
  {country: 'BG', name: 'tapxyhh', points: 35.01, demon: 'UNKNOWN', globalRank: 1221},
  {country: 'RU', name: 'Leeya', points: 12.63, demon: 'Arctic Lights', globalRank: 1858},
  {country: 'RU', name: 'kocheryzhka', points: 11.99, demon: 'Bloodlust', globalRank: 1899},
  {country: 'BY', name: 'ramp1941', points: 11.34, demon: 'shimmer', globalRank: 1961},
  {country: 'RU', name: 'samoletik', points: 10.63, demon: 'Sink', globalRank: 2022},
  {country: 'RU', name: 'vv4zd', points: 9.98, demon: 'Renevant', globalRank: 2078},
  {country: 'DE', name: 'yeahme', points: 7.10, demon: 'Sonic Wave', globalRank: 2450},
  {country: 'UA', name: 'Vakum', points: 7.07, demon: 'Cobwebs', globalRank: 2457},
  {country: 'RU', name: 'Linqwq', points: 6.52, demon: 'Wasureta', globalRank: 2543},
  {country: 'RU', name: 'H30n41k_GmD', points: 6.46, demon: 'Sink', globalRank: 2561},
  {country: 'AM', name: 'SerGio', points: 6.33, demon: 'Wasureta', globalRank: 2596},
  {country: 'RU', name: 'Спини', points: 5.85, demon: 'RUST', globalRank: 2690},
  {country: 'RU', name: 'RossceorpGD', points: 5.70, demon: 'ZAPHKIEL', globalRank: 2736},
  {country: 'KZ', name: '69liqu69', points: 5.10, demon: 'Molten Core', globalRank: 2880},
  {country: 'UA', name: 'Imdrinkingtea', points: 4.82, demon: 'Quantum Processing', globalRank: 2969},
  {country: 'UA', name: 'dugen', points: 4.53, demon: 'Ghoul', globalRank: 3071},
  {country: 'RU', name: 'kotacub', points: 4.40, demon: 'Golden Club', globalRank: 3124},
  {country: 'RU', name: 'KotKartofel', points: 4.32, demon: 'Pulsar', globalRank: 3151},
  {country: 'RU', name: 'NopanicGD', points: 3.52, demon: 'Diamond Disco', globalRank: 3512},
  {country: 'RU', name: 'NatrixGMD', points: 3.06, demon: 'Congregation', globalRank: 3808},
  {country: 'RS', name: 'zerrga', points: 2.24, demon: 'Quantum Processing', globalRank: 4467},
  {country: 'RU', name: 'paradoxiz', points: 1.82, demon: 'INNARDS', globalRank: 4940},
  {country: 'RU', name: 'toxatort', points: 1.72, demon: 'Blade of Justice', globalRank: 5055},
  {country: 'UA', name: 'fottex', points: 1.47, demon: 'Sonic Wave', globalRank: 5439},
  {country: 'RU', name: 'Marzyiiik', points: 1.19, demon: 'Blade of Justice', globalRank: 6003},
  {country: 'UA', name: 'Daggit', points: 1.02, demon: 'Overtime', globalRank: 6359},
  {country: 'UA', name: 'KasaneTeto', points: 1.01, demon: 'Heavens Door', globalRank: 6377},
  {country: 'RU', name: 'itzslxnq', points: 0.71, demon: 'Shinigami', globalRank: 7500},
  {country: 'RU', name: 'aerongmd', points: 0.63, demon: 'Bloodbath', globalRank: 8042},
  {country: 'RU', name: 'Заварррка', points: 0.58, demon: 'Anahita', globalRank: 8372},
  {country: 'RU', name: 'matveypro13', points: 0.36, demon: 'Hopping Over Puddles', globalRank: 9608},
  {country: 'RU', name: 'Roflin', points: 0.31, demon: 'Cataclysm', globalRank: 9848},
  {country: 'RU', name: 'Filkoty', points: 0.13, demon: 'Me Lin A', globalRank: 11310},
  {country: 'BY', name: 'Denchis', points: 0.13, demon: 'Cataclysm', globalRank: 11473},
  {country: 'RU', name: 'Fanim59', points: 0.12, demon: 'Make It Drop', globalRank: 11787},
  {country: 'UA', name: 'prostoymofficial', points: 0.07, demon: 'Acu', globalRank: 12308},
  {country: 'RU', name: 'DarBeast', points: 0.07, demon: 'Acu', globalRank: 12328},
  {country: 'OTHER', name: 'CharaGMDq', points: 0.00, demon: '—', globalRank: 0},
]

const MOCK_EVENTS = [
  ['axsCwtl2Hro', 'SMLT beats Back on Track', 'beat'], ['58ItHCC8ZKI', 'SMLT beats Every End', 'beat'], ['8RduRoZmFHs', 'SMLT beats Ne Rofl Collab', 'beat'],
  ['dWnJw60mig0', 'SMLT beats Society', 'beat'], ['Jiby_l51wpA', 'SMLT beats The Golden', 'beat'],
  ['ThvJhbEtvyQ', 'SMLT beats Firework', 'beat'], ['EBGoVlwTieA', 'SMLT beats Tidalbaeb', 'beat'],
  ['AL-c39hXPcU', '[SMLT] Clutter', 'project'], ['KSHADY9mD4o', '[SMLT] PACMAN', 'project'],
  ['hxFDFHR-U_8', '[SMLT] Hopes and Dream', 'project'], ['jkxZe_CPQFQ', '[SMLT] Caterpillar Blitz', 'project'],
  ['dTHVXborVWs', '[SMLT] Rumn Bass', 'project'], ['aTxro5xLt64', '[SMLT] METPO', 'project'],
  ['KErAay1xmRY', '[SMLT] Parfait', 'project'], ['1HsFMyBgvZU', '[SMLT] C TObOY', 'project'],
  ['uPcrGc1Gn1Q', '[SMLT] Jack', 'project'], ['D3AIJIanoTo', '[SMLT] Ne Rofl Collab', 'project'],
  ['0fKT9D63UFE', '[SMLT] Nuclear Fusion', 'project'], ['7UcYhRcM6e4', '[SMLT] Pronyx Code', 'project'],
  ['9jXskpaV4PI', 'Nuclear Fusion by SMLT', 'project'], ['REHYPVCCS_A', '[SMLT] Nuclear Fusion update', 'project'],
  ['mdAmi4JTGjo', '[SMLT] Nuclear Fusion', 'project'], ['wmzNH3Ln9WU', '[SMLT] METRO', 'project'],
]

const CAPTCHA_WORDS = ['ПАРОЛЬ', 'ДЕМОН', 'РЕЙТИНГ', 'ОЧКИ', 'ИГРОК', 'SMLT', 'ADMIN', 'СЕРВЕР']

// Dev-only API mock so `pnpm dev` works without the Go backend + PostgreSQL.
// Handles /api/* entirely in memory. Never runs in production builds.
export default function devApiMock() {
  const players = MOCK_PLAYERS.map((p, i) => ({id: i + 1, rank: i + 1, ...p}))
  const events = MOCK_EVENTS.map(([videoId, title, category], i) => ({id: i + 1, videoId, title, category, sortOrder: i}))
  const captchas = new Map()
  const changes = [{
    id: 1,
    playerName: 'npoctou_gamer',
    oldRank: 6,
    newRank: 4,
    abovePlayer: 'Flik',
    belowPlayer: 'Tikys',
    passedPlayers: ['Tikys'],
    createdAt: new Date().toISOString(),
  }]
  let nextPlayerId = players.length + 1
  let nextEventId = events.length + 1
  let nextChangeId = 2
  let nextCaptchaId = 1

  const json = (res, data, status = 200) => {
    res.statusCode = status
    res.setHeader('Content-Type', 'application/json')
    res.end(JSON.stringify(data))
  }
  const fail = (res, message, status = 400) => json(res, {success: false, message}, status)
  const failAuth = res => fail(res, 'Не авторизован', 401)
  const isAuthed = req => Boolean(req.headers.cookie && req.headers.cookie.includes('smlt_session=mock'))

  function readBody(req) {
    return new Promise((resolve, reject) => {
      let body = ''
      req.on('data', chunk => { body += chunk; if (body.length > 1e6) req.destroy() })
      req.on('end', () => { try { resolve(body ? JSON.parse(body) : {}) } catch (err) { reject(err) } })
      req.on('error', reject)
    })
  }

  function refreshRanks() {
    players.sort((a, b) => Number(b.points) - Number(a.points) || a.name.localeCompare(b.name))
    players.forEach((p, i) => { p.rank = i + 1 })
  }

  function recordChange(before, after) {
    if (before.rank === after.rank) return
    const passed = before.rank > after.rank ? before.slice(after.rank, before.rank).map(p => p.name) : []
    changes.unshift({
      id: nextChangeId++,
      playerName: after.name,
      oldRank: before.rank,
      newRank: after.rank,
      abovePlayer: after[after.rank - 2]?.name || '',
      belowPlayer: after[after.rank]?.name || '',
      passedPlayers: passed,
      createdAt: new Date().toISOString(),
    })
  }

  function captchaImage(id) {
    const word = captchas.get(String(id))
    const text = word || 'ERR'
    return `<svg xmlns="http://www.w3.org/2000/svg" width="240" height="74" viewBox="0 0 240 74"><rect width="240" height="74" fill="#ffffff"/><circle cx="35" cy="20" r="8" fill="#d8dee9" opacity=".8"/><circle cx="200" cy="46" r="11" fill="#cfe3ff" opacity=".9"/><line x1="10" y1="60" x2="230" y2="18" stroke="#dce1ea" stroke-width="2"/><text x="120" y="50" text-anchor="middle" font-family="monospace" font-size="34" font-weight="bold" fill="#1b2733" transform="rotate(-4 120 50)">${text}</text></svg>`
  }

  return {
    name: 'smlt-dev-api-mock',
    apply: 'serve',
    configureServer(server) {
      server.middlewares.use('/api', async (req, res, next) => {
        const url = new URL(req.url, 'http://localhost')
        const path = url.pathname.replace(/^\/api/, '')

        try {
          if (req.method === 'GET' && path === '/players') return json(res, {success: true, players})

          if (req.method === 'POST' && path === '/players') {
            if (!isAuthed(req)) return failAuth(res)
            const body = await readBody(req)
            const existing = players.find(p => p.name.toLowerCase() === String(body.name || '').toLowerCase())
            if (existing || !body.name) return fail(res, existing ? 'Игрок уже существует' : 'Имя обязательно')
            const player = {
              id: nextPlayerId++,
              name: String(body.name).trim(),
              country: String(body.country || 'OTHER').toUpperCase(),
              points: Number(body.points || 0),
              demon: String(body.demon || '—'),
              globalRank: Number(body.globalRank || 0),
            }
            players.push(player)
            const before = players.map(p => ({...p}))
            refreshRanks()
            recordChange(before.find(p => p.id === player.id), players)
            return json(res, {success: true, message: 'Игрок добавлен'})
          }

          if (req.method === 'DELETE' && path.startsWith('/players/')) {
            if (!isAuthed(req)) return failAuth(res)
            const name = decodeURIComponent(path.slice('/players/'.length)).replace(/\+/g, ' ')
            const index = players.findIndex(p => p.name.toLowerCase() === name.toLowerCase())
            if (index < 0) return fail(res, 'Игрок не найден', 404)
            players.splice(index, 1)
            refreshRanks()
            return json(res, {success: true, message: 'Игрок удалён'})
          }

          if (req.method === 'PUT' && path.startsWith('/players/')) {
            if (!isAuthed(req)) return failAuth(res)
            const oldName = decodeURIComponent(path.slice('/players/'.length)).replace(/\+/g, ' ')
            const body = await readBody(req)
            const index = players.findIndex(p => p.name.toLowerCase() === oldName.toLowerCase())
            if (index < 0) return fail(res, 'Игрок не найден', 404)
            const before = players.map(p => ({...p}))
            players[index] = {...players[index], ...{name: body.name, country: body.country, points: Number(body.points), demon: body.demon, globalRank: Number(body.globalRank)}}
            refreshRanks()
            recordChange(before[index], players)
            return json(res, {success: true, message: 'Игрок обновлён'})
          }

          if (req.method === 'GET' && path === '/search-demonlist') {
            const name = String(url.searchParams.get('name') || '').trim()
            if (!name) return fail(res, 'Введите имя', 400)
            if (players.some(p => p.name.toLowerCase() === name.toLowerCase())) return fail(res, 'Игрок уже в рейтинге', 404)
            const hash = [...name].reduce((sum, ch) => sum + ch.charCodeAt(0), 0)
            return json(res, {success: true, name, country: 'RU', points: Number(((hash % 900) + 1)).toFixed(2), demon: 'Unknown Demon', globalRank: 1000 + (hash % 2500)})
          }

          if (req.method === 'GET' && path === '/recent-changes') return json(res, {success: true, changes})

          if (req.method === 'GET' && path === '/player-levels') {
            const id = String(url.searchParams.get('id') || '').trim()
            const name = String(url.searchParams.get('name') || '').trim()
            if (!id && !name) return fail(res, 'Не указан игрок')
            const t0 = Date.now()
            try {
                const headers = {'User-Agent': 'SMLT-Leaderboard/1.0', Accept: 'application/json'}
                const signal = AbortSignal.timeout(12000)
                let userId = id
                if (!userId && name) {
                    const searchRes = await fetch(`https://api.demonlist.org/leaderboard/user/list?limit=5&search=${encodeURIComponent(name)}`, {headers, signal})
                    const searchData = await searchRes.json()
                    const users = (searchData.data && searchData.data.users) || []
                    const found = users.find(u => u.username.toLowerCase() === name.toLowerCase()) || users[0]
                    if (!found) return json(res, {success: false, message: 'Игрок не найден на demonlist.org'}, 404)
                    userId = String(found.id)
                }
                const upstream = await fetch(`https://api.demonlist.org/user/get?id=${encodeURIComponent(userId)}`, {headers, signal})
                if (!upstream.ok) return json(res, {success: false, message: 'Игрок не найден на demonlist.org'}, 404)
                const data = await upstream.json()
                const info = data.data || {}
                console.log(`[mock:player-levels] id=${id} name="${name}" resolved to user=${userId} in ${Date.now() - t0}ms`)
                return json(res, {success: true, id: userId, username: info.username, placement: info.placement, levels: info.levels || {}})
            } catch (err) {
                if (err.name === 'TimeoutError' || err.name === 'AbortError') return fail(res, 'Demonlist долго не отвечает. Попробуйте ещё раз.', 504)
                return fail(res, 'Не удалось подключиться к demonlist.org', 502)
            }
          }

          if (req.method === 'GET' && path === '/events') return json(res, {success: true, events})

          if (req.method === 'POST' && path === '/events') {
            if (!isAuthed(req)) return failAuth(res)
            const body = await readBody(req)
            events.push({id: nextEventId++, videoId: String(body.videoId), title: String(body.title), category: body.category === 'beat' ? 'beat' : 'project', sortOrder: events.length})
            return json(res, {success: true})
          }

          if (req.method === 'PUT' && path === '/events') {
            if (!isAuthed(req)) return failAuth(res)
            const body = await readBody(req)
            const ids = Array.isArray(body.ids) ? body.ids : []
            ids.forEach((id, order) => {
              const event = events.find(e => e.id === id)
              if (event) event.sortOrder = order
            })
            return json(res, {success: true})
          }

          if (req.method === 'DELETE' && path.startsWith('/events/')) {
            if (!isAuthed(req)) return failAuth(res)
            const id = Number(path.slice('/events/'.length))
            const index = events.findIndex(e => e.id === id)
            if (index < 0) return fail(res, 'Ивент не найден', 404)
            events.splice(index, 1)
            return json(res, {success: true})
          }

          if (req.method === 'GET' && path === '/captcha') {
            const id = 'dev-' + (nextCaptchaId++)
            const word = CAPTCHA_WORDS[Math.floor(Math.random() * CAPTCHA_WORDS.length)]
            captchas.set(id, word)
            return json(res, {success: true, captchaId: id, imageUrl: '/api/captcha/image/' + id})
          }

          if (req.method === 'GET' && path.startsWith('/captcha/image/')) {
            const id = path.slice('/captcha/image/'.length)
            res.statusCode = 200
            res.setHeader('Content-Type', 'image/svg+xml')
            res.setHeader('Cache-Control', 'no-store')
            return res.end(captchaImage(id))
          }

          if (req.method === 'POST' && path === '/auth') {
            const body = await readBody(req)
            const word = captchas.get(String(body.captchaId || ''))
            if (!word) return fail(res, 'Капча истекла — обновите картинку', 400)
            if (!body.captchaAnswer || String(body.captchaAnswer).trim().toLowerCase() !== word.toLowerCase()) return fail(res, 'Неверный ответ на капчу')
            if (!body.password) return fail(res, 'Введите пароль')
            captchas.delete(String(body.captchaId))
            res.setHeader('Set-Cookie', 'smlt_session=mock; Path=/; HttpOnly; Max-Age=86400')
            return json(res, {success: true, token: 'dev-mock-token'})
          }

          if (req.method === 'GET' && path === '/session') {
            if (!isAuthed(req)) return failAuth(res)
            return json(res, {success: true, authenticated: true})
          }

          if (req.method === 'POST' && path === '/logout') {
            res.setHeader('Set-Cookie', 'smlt_session=; Path=/; HttpOnly; Max-Age=0')
            return json(res, {success: true})
          }

          return fail(res, 'Not found in dev mock: ' + req.method + ' /api' + path, 404)
        } catch (err) {
          return fail(res, 'Mock error: ' + err.message, 500)
        }
      })
    },
  }
}