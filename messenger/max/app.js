const $ = id => document.getElementById(id)
const avatarIds = Array.from({length: 20}, (_, i) => `avatar${i + 1}`)
const colors = {violet:'#b9a7ff', blue:'#76b9ff', green:'#75d6a2', yellow:'#f3ca68', orange:'#f3a16b', pink:'#f39ac4', cyan:'#6edbd6', red:'#ff8d9b'}
const query = new URLSearchParams(location.search)
let profile = {avatar_id: 'avatar1', color_id: 'violet'}
let account = null

try { profile = {...profile, ...JSON.parse(localStorage.getItem('maxProfile') || '{}')} } catch {}
if (!avatarIds.includes(profile.avatar_id)) profile.avatar_id = 'avatar1'
if (!colors[profile.color_id]) profile.color_id = 'violet'

function show(id) {
  for (const panel of ['authPanel', 'recoveryDisplay', 'profileEditor']) $(panel).hidden = panel !== id
}

function showAuthForm(id) {
  for (const form of ['loginForm', 'registerForm', 'recoveryForm']) $(form).hidden = form !== id
  $('showLogin').classList.toggle('active', id === 'loginForm')
  $('showRegister').classList.toggle('active', id === 'registerForm')
  $('showLogin').setAttribute('aria-pressed', String(id === 'loginForm'))
  $('showRegister').setAttribute('aria-pressed', String(id === 'registerForm'))
  $('authError').hidden = true
}

function errorMessage(data, fallback) {
  return data && typeof data.error === 'string' ? data.error : fallback
}

async function request(path, payload) {
  const response = await fetch(path, {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(payload),
    credentials: 'same-origin',
  })
  const data = await response.json().catch(() => ({}))
  if (!response.ok) throw new Error(errorMessage(data, 'Не удалось выполнить запрос.'))
  return data
}

function setAccount(next) {
  account = next
  profile = {nickname: next.nickname, avatar_id: next.avatar_id, color_id: next.color_id}
  if (query.get('profile_nickname')) profile.nickname = query.get('profile_nickname').trim().slice(0, 32)
  if (avatarIds.includes(query.get('profile_avatar'))) profile.avatar_id = query.get('profile_avatar')
  if (colors[query.get('profile_color')]) profile.color_id = query.get('profile_color')
  $('nickname').value = profile.nickname
  renderPickers()
}

function renderPickers() {
  const avatars = $('avatars')
  avatars.replaceChildren()
  for (const id of avatarIds) {
    const button = document.createElement('button')
    button.type = 'button'
    button.className = profile.avatar_id === id ? 'selected' : ''
    button.title = id
    button.setAttribute('aria-label', id)
    const image = document.createElement('img')
    image.src = `/avatars/${id}.webp?v=2`
    image.alt = ''
    button.append(image)
    button.onclick = () => { profile.avatar_id = id; renderPickers() }
    avatars.append(button)
  }

  const colorPicker = $('colors')
  colorPicker.replaceChildren()
  for (const [id, value] of Object.entries(colors)) {
    const button = document.createElement('button')
    button.type = 'button'
    button.style.background = value
    button.className = profile.color_id === id ? 'selected' : ''
    button.title = id
    button.setAttribute('aria-label', id)
    button.onclick = () => { profile.color_id = id; renderPickers() }
    colorPicker.append(button)
  }
}

function displayRecoveryCode(code) {
  $('recoveryCodeValue').textContent = code
  show('recoveryDisplay')
}

$('showLogin').onclick = () => { show('authPanel'); showAuthForm('loginForm') }
$('showRegister').onclick = () => { show('authPanel'); showAuthForm('registerForm') }
$('showRecovery').onclick = () => { show('authPanel'); showAuthForm('recoveryForm') }

$('loginForm').onsubmit = async event => {
  event.preventDefault()
  $('authError').hidden = true
  try {
    const data = await request('/api/auth/login', {nickname: $('loginNickname').value, password: $('loginPassword').value})
    setAccount(data.account)
    show('profileEditor')
  } catch (error) {
    $('authError').textContent = error.message
    $('authError').hidden = false
  }
}

$('registerForm').onsubmit = async event => {
  event.preventDefault()
  $('authError').hidden = true
  const password = $('registerPassword').value
  if (password !== $('registerPasswordConfirm').value) {
    $('authError').textContent = 'Пароли не совпадают.'
    $('authError').hidden = false
    return
  }
  try {
    const data = await request('/api/auth/register', {nickname: $('registerNickname').value, password})
    setAccount(data.account)
    displayRecoveryCode(data.recovery_code)
    $('registerPassword').value = ''
    $('registerPasswordConfirm').value = ''
  } catch (error) {
    $('authError').textContent = error.message
    $('authError').hidden = false
  }
}

$('recoveryForm').onsubmit = async event => {
  event.preventDefault()
  $('authError').hidden = true
  try {
    const data = await request('/api/auth/recover', {
      nickname: $('recoveryNickname').value,
      recovery_code: $('recoveryCode').value,
      password: $('recoveryPassword').value,
    })
    setAccount(data.account)
    displayRecoveryCode(data.recovery_code)
    $('recoveryPassword').value = ''
  } catch (error) {
    $('authError').textContent = error.message
    $('authError').hidden = false
  }
}

$('copyRecovery').onclick = async () => {
  try {
    await navigator.clipboard.writeText($('recoveryCodeValue').textContent)
    $('copyRecovery').textContent = 'Скопировано'
  } catch {
    $('copyRecovery').textContent = 'Выдели и скопируй код вручную'
  }
}

$('savedRecovery').onclick = () => show('profileEditor')

$('continue').onclick = async () => {
  const nickname = $('nickname').value.trim()
  if (nickname.length < 3 || nickname.length > 32) {
    $('profileError').textContent = 'Ник должен содержать от 3 до 32 символов.'
    $('profileError').hidden = false
    return
  }
  $('profileError').hidden = true
  $('continue').disabled = true
  try {
    const data = await request('/api/auth/profile', {nickname, avatar_id: profile.avatar_id, color_id: profile.color_id})
    setAccount(data.account)
    localStorage.setItem('maxProfile', JSON.stringify(profile))
    const destination = new URL('https://chat.smlt.lol/')
    destination.searchParams.set('profile_nickname', profile.nickname)
    destination.searchParams.set('profile_avatar', profile.avatar_id)
    destination.searchParams.set('profile_color', profile.color_id)
    if (query.get('room')) destination.searchParams.set('room', query.get('room'))
    location.assign(destination.toString())
  } catch (error) {
    $('profileError').textContent = error.message
    $('profileError').hidden = false
    $('continue').disabled = false
  }
}

function openLogoutConfirmation() {
  $('logoutDialog').showModal()
}

$('cancelLogout').onclick = () => $('logoutDialog').close()
$('logout').onclick = openLogoutConfirmation

$('confirmLogout').onclick = async () => {
  $('logoutDialog').close()
  try {
    await request('/api/auth/logout', {})
    account = null
    show('authPanel')
    showAuthForm('loginForm')
  } catch (error) {
    $('profileError').textContent = error.message
    $('profileError').hidden = false
  }
}

async function init() {
  show('authPanel')
  showAuthForm('loginForm')
  $('loginNickname').value = query.get('profile_nickname') || ''
  $('registerNickname').value = query.get('profile_nickname') || ''
  $('recoveryNickname').value = query.get('profile_nickname') || ''
  try {
    const response = await fetch('/api/auth/me', {credentials: 'same-origin', cache: 'no-store'})
    if (!response.ok) return
    const data = await response.json()
    if (data.authenticated && data.account) {
      setAccount(data.account)
      show('profileEditor')
    }
  } catch {
    $('authError').textContent = 'Не удалось подключиться к серверу. Попробуй обновить страницу.'
    $('authError').hidden = false
  }
}

init()
