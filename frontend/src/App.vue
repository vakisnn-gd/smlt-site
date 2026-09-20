<script setup>
import { onMounted, ref } from 'vue'
import { api } from './api'
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
const lang = ref(localStorage.getItem('smlt-lang') === 'en' ? 'en' : 'ru')
const roughMode = ref(localStorage.getItem('smlt-style') === 'rough')
const isAdmin = ref(false)
const showLogin = ref(false)
const editingPlayer = ref(null)
const showPlayerModal = ref(false)
const notice = ref('')

function viewFromPath() {
  if (location.pathname === '/events') return 'events'
  if (location.pathname === '/about') return 'about'
  return 'leaderboard'
}
const view = ref(viewFromPath())

function applyStyle() {
  document.documentElement.dataset.style = roughMode.value ? 'rough' : 'classic'
}

applyStyle()

function toggleStyle() {
  roughMode.value = !roughMode.value
  localStorage.setItem('smlt-style', roughMode.value ? 'rough' : 'classic')
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
    events: lang.value === 'en' ? 'SMLT collaborations, events and community projects.' : 'Коллабы, ивенты и проекты сообщества SMLT.',
    about: lang.value === 'en' ? 'Information about the SMLT community, contacts and projects.' : 'Информация о сообществе SMLT, контакты и проекты.',
  }
  document.title = titles[view.value]
  const description = document.querySelector('meta[name="description"]')
  if (description) description.setAttribute('content', descriptions[view.value])
}

function navigate(next) {
  view.value = next
  const path = next === 'events' ? '/events' : next === 'about' ? '/about' : '/'
  if (location.pathname !== path) history.pushState({view: next}, '', path)
  updatePageMeta()
  window.scrollTo({top: 0, behavior: 'smooth'})
}

function toggleLanguage() {
  lang.value = lang.value === 'ru' ? 'en' : 'ru'
  localStorage.setItem('smlt-lang', lang.value)
  document.documentElement.lang = lang.value
  updatePageMeta()
}

async function loadData() {
  loading.value = true
  error.value = ''
  try {
    const [playerData, changeData, eventData] = await Promise.all([api.players(), api.changes(), api.events()])
    players.value = playerData.players || []
    changes.value = changeData.changes || []
    events.value = eventData.events || []
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

async function checkSession() {
  try {
    await api.session()
    isAdmin.value = true
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
    await loadData()
  } catch (err) {
    notice.value = err.message
    if (err.status === 401) isAdmin.value = false
  }
}

async function savedPlayer() {
  showPlayerModal.value = false
  notice.value = lang.value === 'en' ? 'Leaderboard updated' : 'Рейтинг обновлён'
  await loadData()
}

async function logout() {
  try { await api.logout() } catch {}
  isAdmin.value = false
}

async function reloadEvents() {
  try {
    const eventData = await api.events()
    events.value = eventData.events || []
    notice.value = lang.value === 'en' ? 'Events updated' : 'Ивенты обновлены'
  } catch (err) { notice.value = err.message }
}

onMounted(() => {
  applyStyle()
  document.documentElement.lang = lang.value
  updatePageMeta()
  loadData()
  checkSession()
  window.addEventListener('popstate', () => { view.value = viewFromPath(); updatePageMeta() })
})
</script>

<template>
  <div class="app-shell">
    <AppHeader :view="view" :lang="lang" :is-admin="isAdmin" :rough-mode="roughMode" @navigate="navigate" @language="toggleLanguage" @style="toggleStyle" @login="showLogin = true" @add="openAdd" @logout="logout" />
    <main>
      <LeaderboardView v-if="view === 'leaderboard'" :players="players" :changes="changes" :loading="loading" :error="error" :lang="lang" :is-admin="isAdmin" @edit="openEdit" @delete="removePlayer" @retry="loadData" />
      <EventsView v-else-if="view === 'events'" :lang="lang" :events="events" :is-admin="isAdmin" @changed="reloadEvents" />
      <AboutView v-else :lang="lang" :is-admin="isAdmin" @leaderboard="navigate('leaderboard')" @events="navigate('events')" @login="showLogin = true" @logout="logout" />
    </main>
    <transition name="toast"><div v-if="notice" class="toast" @click="notice = ''">{{ notice }}</div></transition>
    <LoginModal v-if="showLogin" :lang="lang" @close="showLogin = false" @authenticated="showLogin = false; isAdmin = true" />
    <PlayerModal v-if="showPlayerModal" :lang="lang" :player="editingPlayer" @close="showPlayerModal = false" @saved="savedPlayer" @unauthorized="isAdmin = false" />
  </div>
</template>
