<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from './api'
import { setRoughFlags } from './data'
import AppHeader from './components/AppHeader.vue'
import LeaderboardView from './components/LeaderboardView.vue'
import EventsView from './components/EventsView.vue'
import AboutView from './components/AboutView.vue'
import LoginModal from './components/LoginModal.vue'
import PlayerModal from './components/PlayerModal.vue'

const players = ref([])
const changes = ref([])
const events = ref([])
const loading = ref(true)
const error = ref('')
const lang = ref(location.pathname === '/en' || location.pathname.startsWith('/en/') ? 'en' : 'ru')
const roughMode = ref(localStorage.getItem('smlt-style') === 'rough')
setRoughFlags(roughMode.value)
const isAdmin = ref(false)
const showLogin = ref(false)
const editingPlayer = ref(null)
const showPlayerModal = ref(false)
const notice = ref('')
const showEasterEgg = ref(false)
const route = useRoute()
const router = useRouter()
const loaded = ref({leaderboard: false, events: false})

const view = computed(() => route.meta.view || 'leaderboard')
const playerId = computed(() => String(route.name).endsWith('player') ? String(route.params.id) : '')

function routeName(name, language = lang.value) {
  return language === 'en' ? `en-${name}` : name
}

function applyStyle() {
  document.documentElement.dataset.style = roughMode.value ? 'rough' : 'classic'
}

applyStyle()

function toggleStyle() {
  roughMode.value = !roughMode.value
  localStorage.setItem('smlt-style', roughMode.value ? 'rough' : 'classic')
  setRoughFlags(roughMode.value)
  applyStyle()
}

function updatePageMeta() {
  const titles = {
    leaderboard: lang.value === 'en' ? 'SMLT Leaderboard' : 'SMLT — рейтинг',
    events: lang.value === 'en' ? 'SMLT Events' : 'SMLT — ивенты',
    about: lang.value === 'en' ? 'About SMLT' : 'О SMLT',
  }
  const descriptions = {
    leaderboard: lang.value === 'en' ? 'SMLT community leaderboard and player achievements.' : 'Рейтинг и достижения игроков сообщества SMLT.',
    events: lang.value === 'en' ? 'SMLT events and collaborations.' : 'Ивенты и коллабы сообщества SMLT.',
    about: lang.value === 'en' ? 'Information about the SMLT community, contacts and projects.' : 'Информация о сообществе SMLT, контакты и проекты.',
  }
  const selectedPlayer = playerId.value && players.value.find(player => String(player.gdlId || player.id) === playerId.value)
  document.title = selectedPlayer ? `${selectedPlayer.name} — SMLT` : titles[view.value]
  const description = document.querySelector('meta[name="description"]')
  if (description) description.setAttribute('content', descriptions[view.value])
  const canonicalUrl = `https://smlt.lol${route.path}`
  const canonical = document.querySelector('link[rel="canonical"]')
  if (canonical) canonical.setAttribute('href', canonicalUrl)
  const ogTitle = document.querySelector('meta[property="og:title"]')
  const ogDescription = document.querySelector('meta[property="og:description"]')
  const ogUrl = document.querySelector('meta[property="og:url"]')
  if (ogTitle) ogTitle.setAttribute('content', document.title)
  if (ogDescription) ogDescription.setAttribute('content', descriptions[view.value])
  if (ogUrl) ogUrl.setAttribute('content', canonicalUrl)
  const alternatePath = route.path.replace(/^\/en(?=\/|$)/, '') || '/'
  const ruUrl = `https://smlt.lol${alternatePath}`
  const enUrl = `https://smlt.lol/en${alternatePath === '/' ? '' : alternatePath}`
  document.querySelector('link[hreflang="ru"]')?.setAttribute('href', ruUrl)
  document.querySelector('link[hreflang="en"]')?.setAttribute('href', enUrl)
  document.querySelector('link[hreflang="x-default"]')?.setAttribute('href', ruUrl)
}

function navigate(next) {
  router.push({name: routeName(next)})
}

function toggleLanguage() {
  const next = lang.value === 'ru' ? 'en' : 'ru'
  localStorage.setItem('smlt-lang', next)
  const target = playerId.value ? 'player' : view.value
  router.push({name: routeName(target, next), params: playerId.value ? {id: playerId.value} : {}})
}

async function loadLeaderboard(force = false) {
  if (loaded.value.leaderboard && !force) return
  loading.value = true
  error.value = ''
  try {
    const results = await Promise.allSettled([api.players(), api.changes()])
    if (results[0].status === 'rejected') throw results[0].reason
    players.value = results[0].value.players || []
    if (results[1].status === 'fulfilled') changes.value = results[1].value.changes || []
    loaded.value.leaderboard = true
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

async function loadEvents(force = false) {
  if (loaded.value.events && !force) return true
  try {
    const eventData = await api.events()
    events.value = eventData.events || []
    loaded.value.events = true
    return true
  } catch (err) {
    notice.value = err.message
    return false
  }
}

function loadRouteData() {
  if (view.value === 'leaderboard') loadLeaderboard()
  if (view.value === 'events') loadEvents()
}

async function checkSession() {
  try {
    const session = await api.session()
    isAdmin.value = session.authenticated === true
  } catch {
    isAdmin.value = false
  }
}

function openAdd() {
  editingPlayer.value = null
  showPlayerModal.value = true
}

function openEdit(player) {
  editingPlayer.value = player
  showPlayerModal.value = true
}

async function removePlayer(player) {
  const message = lang.value === 'en' ? `Delete ${player.name}?` : `Удалить игрока ${player.name}?`
  if (!confirm(message)) return
  try {
    await api.deletePlayer(player.name)
    notice.value = lang.value === 'en' ? 'Player deleted' : 'Игрок удалён'
    await loadLeaderboard(true)
  } catch (err) {
    notice.value = err.message
    if (err.status === 401) isAdmin.value = false
  }
}

async function removeChange(change) {
  if (!change?.eventType) return
  const message = lang.value === 'en' ? 'Delete this history entry?' : 'Удалить эту запись из истории?'
  if (!confirm(message)) return
  try {
    await api.deletePlayerEvent(change.id)
    notice.value = lang.value === 'en' ? 'History entry deleted' : 'Запись удалена из истории'
    await loadLeaderboard(true)
  } catch (err) {
    notice.value = err.message
    if (err.status === 401) isAdmin.value = false
  }
}

async function savedPlayer() {
  showPlayerModal.value = false
  notice.value = lang.value === 'en' ? 'Leaderboard updated' : 'Рейтинг обновлён'
  await loadLeaderboard(true)
}

async function logout() {
  try { await api.logout() } catch {}
  isAdmin.value = false
}

async function reloadEvents() {
  if (await loadEvents(true)) {
    notice.value = lang.value === 'en' ? 'Events updated' : 'Ивенты обновлены'
  }
}

onMounted(() => {
  lang.value = route.meta.lang || 'ru'
  applyStyle()
  document.documentElement.lang = lang.value
  updatePageMeta()
  loadRouteData()
  checkSession()
})

watch(() => route.fullPath, () => {
  lang.value = route.meta.lang || 'ru'
  document.documentElement.lang = lang.value
  loadRouteData()
  updatePageMeta()
})
watch(players, updatePageMeta)
</script>

<template>
  <div class="app-shell">
    <AppHeader :view="view" :lang="lang" :is-admin="isAdmin" :rough-mode="roughMode" @navigate="navigate" @language="toggleLanguage" @style="toggleStyle" @easter-egg="showEasterEgg = true" @login="showLogin = true" @add="openAdd" @logout="logout" />
    <main>
      <LeaderboardView v-if="view === 'leaderboard'" :players="players" :changes="changes" :loading="loading" :error="error" :lang="lang" :is-admin="isAdmin" :player-id="playerId" @edit="openEdit" @delete="removePlayer" @delete-change="removeChange" @retry="loadLeaderboard(true)" @open-player="id => router.push({name: routeName('player'), params: {id}})" @close-player="router.push({name: routeName('leaderboard')})" />
      <EventsView v-else-if="view === 'events'" :lang="lang" :events="events" :is-admin="isAdmin" @changed="reloadEvents" />
      <AboutView v-else :lang="lang" :is-admin="isAdmin" @leaderboard="navigate('leaderboard')" @events="navigate('events')" @login="showLogin = true" @logout="logout" />
    </main>
    <transition name="toast"><div v-if="notice" class="toast" role="status" aria-live="polite" @click="notice = ''">{{ notice }}</div></transition>
    <div v-if="showEasterEgg" class="easter-backdrop">
      <section class="easter-card" role="dialog" aria-modal="true" :aria-label="lang === 'en' ? 'SMLT easter egg' : 'Пасхалка SMLT'" @keydown.esc="showEasterEgg = false">
        <button class="modal-close" :aria-label="lang === 'en' ? 'Close' : 'Закрыть'" @click="showEasterEgg = false">×</button>
        <div class="easter-mark" aria-hidden="true">⚡</div>
        <p class="eyebrow">{{ lang === 'en' ? 'SMLT SECRET MODE' : 'СЕКРЕТНЫЙ РЕЖИМ SMLT' }}</p>
        <h2>{{ lang === 'en' ? 'YOU ARE THE SMLT ADMIN' : 'ВЫ АДМИН SMLT' }}</h2>
        <p>{{ lang === 'en' ? 'Just kidding… or are you?' : 'Шутка… наверное :)' }}</p>
        <button class="primary-button" @click="showEasterEgg = false">{{ lang === 'en' ? 'I knew it' : 'Я так и знал' }}</button>
      </section>
    </div>
    <LoginModal v-if="showLogin" :lang="lang" @close="showLogin = false" @authenticated="showLogin = false; isAdmin = true" />
    <PlayerModal v-if="showPlayerModal" :lang="lang" :player="editingPlayer" @close="showPlayerModal = false" @saved="savedPlayer" @unauthorized="isAdmin = false" />
  </div>
</template>
