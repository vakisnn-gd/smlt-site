<script setup>
import { computed, ref, watch } from 'vue'
import { api } from '../api'

const props = defineProps({lang: String, events: {type: Array, default: () => []}, isAdmin: Boolean})
const emit = defineEmits(['changed'])
const selectedBeat = ref(null)
const selectedProject = ref(null)
const showForm = ref(false)
const form = ref({videoId: '', title: '', category: 'beat', status: 'planned'})
const formError = ref('')
const saving = ref(false)
const beats = computed(() => props.events.filter(event => event.category === 'beat'))
const projects = computed(() => props.events.filter(event => event.category === 'project'))
const collabStatuses = [
  {id: 'ready-verify', ru: 'готов и верифнут', en: 'ready and verified'},
  {id: 'ready', ru: 'готов', en: 'ready'},
  {id: 'in-progress', ru: 'в процессе', en: 'in progress'},
  {id: 'planned', ru: 'планируется', en: 'planned'},
  {id: 'dead', ru: 'мёртв', en: 'dead'},
  {id: 'frozen', ru: 'заморожен', en: 'frozen'},
]

watch(() => props.events, list => {
  if (!selectedBeat.value || !beats.value.some(event => event.id === selectedBeat.value.id)) selectedBeat.value = beats.value[0] || null
  if (!selectedProject.value || !projects.value.some(event => event.id === selectedProject.value.id)) selectedProject.value = projects.value[0] || null
}, {immediate: true})

function videoUrl(event) { return `https://www.youtube-nocookie.com/embed/${event.videoId}?rel=0&modestbranding=1` }
function statusMeta(status) { return collabStatuses.find(item => item.id === status) || collabStatuses[3] }
function statusLabel(status) { const item = statusMeta(status); return props.lang === 'en' ? item.en : item.ru }
function extractYouTubeId(value) {
  const raw = String(value || '').trim()
  if (!raw) return ''
  if (/^[A-Za-z0-9_-]{11}$/.test(raw)) return raw

  let parsed
  try {
    parsed = new URL(/^[a-z][a-z\d+.-]*:\/\//i.test(raw) ? raw : `https://${raw}`)
  } catch {
    return ''
  }

  const host = parsed.hostname.toLowerCase().replace(/^www\./, '')
  let id = ''
  if (host === 'youtu.be') {
    id = parsed.pathname.split('/').filter(Boolean)[0] || ''
  } else if (host === 'youtube.com' || host === 'm.youtube.com' || host === 'music.youtube.com' || host === 'youtube-nocookie.com') {
    if (parsed.pathname === '/watch') id = parsed.searchParams.get('v') || ''
    else {
      const parts = parsed.pathname.split('/').filter(Boolean)
      if (['embed', 'shorts', 'live', 'v'].includes(parts[0])) id = parts[1] || ''
    }
  }
  return /^[A-Za-z0-9_-]{11}$/.test(id) ? id : ''
}

async function addEvent() {
  saving.value = true; formError.value = ''
  const videoId = extractYouTubeId(form.value.videoId)
  if (!videoId) {
    formError.value = props.lang === 'en' ? 'Paste a YouTube link or an 11-character video ID.' : 'Вставьте ссылку YouTube или 11-символьный ID видео.'
    saving.value = false
    return
  }
  try { await api.addEvent({...form.value, videoId}); form.value = {videoId: '', title: '', category: 'beat', status: 'planned'}; showForm.value = false; emit('changed') }
  catch (err) { formError.value = err.message }
  finally { saving.value = false }
}
async function deleteEvent(event) {
  if (!confirm(props.lang === 'en' ? `Delete “${event.title}”?` : `Удалить «${event.title}»?`)) return
  try { await api.deleteEvent(event.id); emit('changed') } catch (err) { formError.value = err.message }
}
async function move(event, direction) {
  const same = props.events.filter(item => item.category === event.category)
  const index = same.findIndex(item => item.id === event.id)
  const target = index + direction
  if (index < 0 || target < 0 || target >= same.length) return
  ;[same[index], same[target]] = [same[target], same[index]]
  const other = props.events.filter(item => item.category !== event.category)
  const list = event.category === 'beat' ? [...same, ...other] : [...other, ...same]
  try { await api.reorderEvents(list.map(item => item.id)); emit('changed') } catch (err) { formError.value = err.message }
}
async function updateStatus(event, status) {
  try { await api.updateEventStatus(event.id, status); emit('changed') } catch (err) { formError.value = err.message }
}
</script>

<template>
  <div class="page events-page">
    <header class="events-heading"><div><h1>{{ lang === 'en' ? 'Events and collabs' : 'Ивенты и коллабы' }}</h1><p>{{ lang === 'en' ? 'Completions and collaborations by SMLT.' : 'Совместные прохождения и коллабы SMLT.' }}</p></div><button v-if="isAdmin" class="secondary-button" @click="showForm = !showForm">{{ showForm ? '×' : '+' }} {{ lang === 'en' ? 'Manage events' : 'Управление' }}</button></header>
    <form v-if="showForm" class="event-admin" @submit.prevent="addEvent">
      <input v-model="form.videoId" :aria-label="lang === 'en' ? 'YouTube link or video ID' : 'Ссылка YouTube или ID видео'" :placeholder="lang === 'en' ? 'YouTube link or video ID' : 'Ссылка YouTube или ID видео'" required maxlength="200" autocomplete="off">
      <input v-model="form.title" :aria-label="lang === 'en' ? 'Title' : 'Название'" :placeholder="lang === 'en' ? 'Title' : 'Название'" required maxlength="120">
      <select v-model="form.category" :aria-label="lang === 'en' ? 'Category' : 'Категория'"><option value="beat">{{ lang === 'en' ? 'SMLT Events' : 'Ивенты' }}</option><option value="project">{{ lang === 'en' ? 'Collabs' : 'Коллабы' }}</option></select>
      <select v-model="form.status" :disabled="form.category !== 'project'" :aria-label="lang === 'en' ? 'Collab status' : 'Статус коллаба'"><option v-for="status in collabStatuses" :key="status.id" :value="status.id">{{ lang === 'en' ? status.en : status.ru }}</option></select>
      <button class="primary-button" :disabled="saving">{{ lang === 'en' ? 'Add' : 'Добавить' }}</button>
      <p v-if="formError" class="form-error">{{ formError }}</p>
    </form>
    <section class="video-section">
      <div class="section-head"><h2>{{ lang === 'en' ? 'SMLT Events' : 'Ивенты' }}</h2></div>
      <div class="video-layout">
        <div class="video-tabs" role="tablist"><div v-for="(event, index) in beats" :key="event.id" class="event-tab-row"><button role="tab" :aria-selected="selectedBeat && selectedBeat.id === event.id" :class="{active: selectedBeat && selectedBeat.id === event.id}" @click="selectedBeat = event">{{ event.title }}</button><template v-if="isAdmin"><button class="event-order" @click="move(event, -1)" :disabled="index === 0">↑</button><button class="event-order" @click="move(event, 1)" :disabled="index === beats.length - 1">↓</button><button class="event-order danger" @click="deleteEvent(event)">×</button></template></div></div>
        <div v-if="selectedBeat" class="video-card"><iframe :src="videoUrl(selectedBeat)" :title="selectedBeat.title" loading="lazy" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe><h3>{{ selectedBeat.title }}</h3></div>
      </div>
    </section>
    <section class="video-section collabs-section">
      <div class="section-head"><h2>{{ lang === 'en' ? 'Collabs' : 'Коллабы' }}</h2></div>
      <div class="collab-status-legend" :aria-label="lang === 'en' ? 'Collab status legend' : 'Обозначения статусов коллабов'">
        <span v-for="status in collabStatuses" :key="status.id" class="collab-status-legend-item"><i :class="['collab-status-square', `status-${status.id}`]" aria-hidden="true"></i>{{ lang === 'en' ? status.en : status.ru }}</span>
      </div>
      <div class="video-layout">
        <div class="video-tabs" role="tablist"><div v-for="(event, index) in projects" :key="event.id" class="event-tab-row"><span :class="['collab-status-square', `status-${statusMeta(event.status).id}`]" role="img" :aria-label="statusLabel(event.status)" :title="statusLabel(event.status)"></span><button role="tab" :aria-selected="selectedProject && selectedProject.id === event.id" :class="{active: selectedProject && selectedProject.id === event.id}" @click="selectedProject = event">{{ event.title }}</button><template v-if="isAdmin"><select class="event-status-select" :aria-label="lang === 'en' ? `Status: ${event.title}` : `Статус: ${event.title}`" :value="statusMeta(event.status).id" @change="updateStatus(event, $event.target.value)"><option v-for="status in collabStatuses" :key="status.id" :value="status.id">{{ lang === 'en' ? status.en : status.ru }}</option></select><button class="event-order" @click="move(event, -1)" :disabled="index === 0">↑</button><button class="event-order" @click="move(event, 1)" :disabled="index === projects.length - 1">↓</button><button class="event-order danger" @click="deleteEvent(event)">×</button></template></div></div>
        <div v-if="selectedProject" class="video-card"><iframe :src="videoUrl(selectedProject)" :title="selectedProject.title" loading="lazy" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe><h3>{{ selectedProject.title }}</h3></div>
      </div>
    </section>
  </div>
</template>
