<script setup>
import { onMounted, ref } from 'vue'
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
const lang = ref(localStorage.getItem('smlt-lang') === 'en' ? 'en' : 'ru')
const roughMode = ref(localStorage.getItem('smlt-style') === 'rough')
setRoughFlags(roughMode.value)
const isAdmin = ref(false)
const showLogin = ref(false)
const editingPlayer = ref(null)
const showPlayerModal = ref(false)
const notice = ref('')
const showEasterEgg = ref(false)

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
    <AppHeader :view="view" :lang="lang" :is-admin="isAdmin" :rough-mode="roughMode" @navigate="navigate" @language="toggleLanguage" @style="toggleStyle" @easter-egg="showEasterEgg = true" @login="showLogin = true" @add="openAdd" @logout="logout" />
    <main>
      <LeaderboardView v-if="view === 'leaderboard'" :players="players" :changes="changes" :loading="loading" :error="error" :lang="lang" :is-admin="isAdmin" @edit="openEdit" @delete="removePlayer" @retry="loadData" />
      <EventsView v-else-if="view === 'events'" :lang="lang" :events="events" :is-admin="isAdmin" @changed="reloadEvents" />
      <AboutView v-else :lang="lang" :is-admin="isAdmin" @leaderboard="navigate('leaderboard')" @events="navigate('events')" @login="showLogin = true" @logout="logout" />
    </main>
    <transition name="toast"><div v-if="notice" class="toast" @click="notice = ''">{{ notice }}</div></transition>
    <div v-if="showEasterEgg" class="easter-backdrop">
      <section class="easter-card" role="dialog" aria-modal="true" :aria-label="lang === 'en' ? 'SMLT easter egg' : 'Пасхалка SMLT'">
        <button class="modal-close" :aria-label="lang === 'en' ? 'Close' : 'Закрыть'" @click="showEasterEgg = false">×</button>
        <div class="easter-mark" aria-hidden="true">⚡</div>
        <p class="eyebrow">{{ lang === 'en' ? 'SMLT SECRET MODE' : 'СЕКРЕТНЫЙ РЕЖИМ SMLT' }}</p>
        <p>{{ lang === 'en' ? 'Just kidding… or are you?' : 'Шутка… наверное :)' }}</p>
        <p class="easter-hint">{{ lang === 'en' ? 'It stays open until you close it. Trigger: five quick clicks on the SMLT logo.' : 'Окно не исчезает само. Пасхалка открывается после пяти быстрых нажатий на логотип SMLT.' }}</p>
        <button class="primary-button" @click="showEasterEgg = false">{{ lang === 'en' ? 'I knew it' : 'Я так и знал' }}</button>
      </section>
    </div>
    <LoginModal v-if="showLogin" :lang="lang" @close="showLogin = false" @authenticated="showLogin = false; isAdmin = true" />
    <PlayerModal v-if="showPlayerModal" :lang="lang" :player="editingPlayer" @close="showPlayerModal = false" @saved="savedPlayer" @unauthorized="isAdmin = false" />
  </div>
</template>
