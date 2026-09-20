<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { api } from '../api'
import { countryMeta } from '../data'

const props = defineProps({player: Object, lang: String, players: Array})
const emit = defineEmits(['close'])

const sorted = computed(() => [...(props.players || [])].sort((a, b) => Number(b.points) - Number(a.points) || a.name.localeCompare(b.name)))
const countryRank = computed(() => {
  const same = sorted.value.filter(p => p.country === props.player.country)
  const pos = same.findIndex(p => p.id === props.player.id)
  return pos + 1
})
const countryTotal = computed(() => sorted.value.filter(p => p.country === props.player.country).length)

const records = ref(null)
const levelsLoading = ref(false)
const levelsError = ref('')
const profileId = computed(() => Number(records.value?.id || props.player.gdlId) || 0)
const requestTimeout = 8000

const sections = computed(() => {
  const l = records.value?.levels
  if (!l) return []
  const defs = [
    {key: 'main', label: 'Main'},
    {key: 'advanced', label: 'Advanced'},
    {key: 'extended', label: 'Extended'},
    {key: 'unbounded', label: 'Unbounded'},
    {key: 'progress', label: props.lang === 'en' ? 'In progress' : 'В процессе'},
    {key: 'verified', label: 'Verified'},
  ]
  const result = []
  for (const def of defs) {
    const items = l[def.key] || []
    if (items.length) result.push({...def, items})
  }
  return result
})

async function load() {
  const playerName = props.player.name
  const gdlId = props.player.gdlId
  records.value = null
  levelsError.value = ''
  levelsLoading.value = true
  let settled = false
  const guard = new Promise(resolve => setTimeout(() => {
    if (!settled) {
      settled = true
      levelsLoading.value = false
      levelsError.value = props.lang === 'en' ? 'Demonlist is taking too long. Try again.' : 'Demonlist долго не отвечает. Попробуйте ещё раз.'
      console.warn('[player-details] timeout for', playerName)
    }
  }, requestTimeout))
  try {
    const data = await Promise.race([api.levels(gdlId, playerName), guard.then(() => null)])
    if (settled) return
    settled = true
    records.value = data
    console.info('[player-details]', playerName, 'levels:', Object.keys(data.levels || {}).join(', ') || 'none')
  } catch (err) {
    if (!settled) {
      settled = true
      levelsError.value = err.message
      console.error('[player-details] error for', playerName, err)
    }
  } finally {
    levelsLoading.value = false
  }
}
watch(() => props.player, load, {immediate: true})

function onKeydown(event) {
  if (event.key === 'Escape') emit('close')
}
onMounted(() => {
  window.addEventListener('keydown', onKeydown)
  document.body.style.overflow = 'hidden'
})
onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown)
  document.body.style.overflow = ''
})
</script>

<template>
  <div class="modal-backdrop" @click.self="$emit('close')">
    <section class="modal player-details-modal" role="dialog" aria-modal="true">
      <button class="modal-close" @click="$emit('close')">×</button>
      <div class="pdetails-hero">
        <img v-if="countryMeta(player.country, lang).flagSrc" class="pdetails-flag" :src="countryMeta(player.country, lang).flagSrc" :alt="countryMeta(player.country, lang).name"><span v-else class="pdetails-flag">{{ countryMeta(player.country, lang).flag }}</span>
        <div>
          <p class="eyebrow">{{ lang === 'en' ? 'SMLT PLAYER' : 'ИГРОК SMLT' }}</p>
          <h2 class="pdetails-name">{{ player.name }}</h2>
          <p class="pdetails-country">{{ countryMeta(player.country, lang).name }} · #{{ player.globalRank || '—' }} {{ lang === 'en' ? 'world' : 'в мире' }}</p>
        </div>
      </div>
      <div class="pdetails-stats">
        <div class="pdetails-stat"><strong>#{{ player.rank }}</strong><small>{{ lang === 'en' ? 'SMLT rank' : 'Место SMLT' }}</small></div>
        <div class="pdetails-stat"><strong>{{ Number(player.points).toFixed(2) }}</strong><small>{{ lang === 'en' ? 'Points' : 'Очки' }}</small></div>
        <div class="pdetails-stat"><strong>#{{ player.globalRank || '—' }}</strong><small>{{ lang === 'en' ? 'World rank' : 'Мировой рейтинг' }}</small></div>
      </div>
      <p class="pdetails-demon">{{ lang === 'en' ? 'Hardest demon' : 'Сложнейший демон' }}: <strong>{{ player.demon || '—' }}</strong></p>
      <div class="pdetails-chips">
        <span class="pdetails-chip">{{ lang === 'en' ? 'Player' : 'Игрок' }} {{ countryRank }} / {{ countryTotal }} {{ lang === 'en' ? 'of' : 'из' }} {{ countryMeta(player.country, lang).name }}</span>
        <a v-if="profileId" class="pdetails-chip pdetails-link" :href="`https://demonlist.org/profile/${profileId}`" target="_blank" rel="noopener">Demonlist ↗</a>
      </div>
      <a v-if="profileId" :href="`https://demonlist.org/profile/${profileId}`" class="primary-button pdetails-open" target="_blank" rel="noopener">{{ lang === 'en' ? 'Open on Demonlist' : 'Открыть на Demonlist' }}</a>
      <div class="pdetails-levels">
        <div class="pdetails-levels-head"><h3>{{ lang === 'en' ? 'Completed levels' : 'Пройденные уровни' }}</h3><span v-if="levelsLoading" class="mini-loader"></span></div>
        <div v-if="levelsLoading" class="pdetails-levels-state">{{ lang === 'en' ? 'Loading levels…' : 'Загружаем уровни…' }}</div>
        <div v-else-if="levelsError" class="pdetails-levels-state pdetails-levels-error"><p class="form-error">{{ levelsError }}</p><button class="secondary-button pdetails-retry" @click="load">{{ lang === 'en' ? 'Try again' : 'Повторить' }}</button></div>
        <template v-else-if="sections.length">
          <section v-for="section in sections" :key="section.key" class="pdetails-subsection">
            <div class="pdetails-subsection-head"><h4>{{ section.label }}</h4><span>{{ section.items.length }}</span></div>
            <ol class="pdetails-level-list">
              <li v-for="item in section.items" :key="item.id">
                <span class="pdetails-lvl-pos">{{ item.placement }}</span>
                <a v-if="item.video_url" :href="item.video_url" target="_blank" rel="noopener">{{ item.name }}</a>
                <span v-else class="pdetails-lvl-name">{{ item.name }}</span>
                <small v-if="item.percent">{{ item.percent }}%</small>
              </li>
            </ol>
          </section>
        </template>
        <p v-else class="pdetails-levels-state">{{ lang === 'en' ? 'No completed levels on Demonlist.' : 'На Demonlist нет пройденных уровней.' }}</p>
      </div>
    </section>
  </div>
</template>
