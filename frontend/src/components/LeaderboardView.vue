<script setup>
import { computed, ref } from 'vue'
import { countryMeta } from '../data'
import RecentChanges from './RecentChanges.vue'
import PlayerDetailsView from './PlayerDetailsView.vue'

const props = defineProps({players: Array, changes: Array, loading: Boolean, error: String, lang: String, isAdmin: Boolean})
defineEmits(['edit', 'delete', 'retry'])
const search = ref('')
const selected = ref(null)
const country = ref('ALL')
const spinPlayer = ref(null)
const spinRolling = ref(false)

const sorted = computed(() => [...(props.players || [])].sort((a, b) => Number(b.points) - Number(a.points) || a.name.localeCompare(b.name)))
const visible = computed(() => sorted.value.filter(player => {
  const matchesSearch = player.name.toLowerCase().includes(search.value.trim().toLowerCase())
  return matchesSearch && (country.value === 'ALL' || player.country === country.value)
}))
const countries = computed(() => [...new Set(sorted.value.map(player => player.country))].sort((a, b) => (a === 'OTHER') - (b === 'OTHER') || a.localeCompare(b)))
const total = computed(() => sorted.value.reduce((sum, player) => sum + Number(player.points || 0), 0))
const average = computed(() => sorted.value.length ? total.value / sorted.value.length : 0)
const countryStats = computed(() => {
  const counts = new Map()
  for (const player of sorted.value) counts.set(player.country, (counts.get(player.country) || 0) + 1)
  return [...counts.entries()].sort((a, b) => (a[0] === 'OTHER') - (b[0] === 'OTHER') || b[1] - a[1])
})

function wait(ms) { return new Promise(resolve => setTimeout(resolve, ms)) }

async function spinSMLT() {
  if (!sorted.value.length || spinRolling.value) return
  spinRolling.value = true
  for (let i = 0; i < 12; i += 1) {
    spinPlayer.value = sorted.value[Math.floor(Math.random() * sorted.value.length)]
    await wait(55 + i * 9)
  }
  spinRolling.value = false
}
</script>

<template>
  <div class="page leaderboard-page">
    <div class="leaderboard-layout">
      <div class="leaderboard-col">
        <section class="leaderboard-card">
          <div class="leaderboard-toolbar">
            <div class="section-kicker"><strong>{{ lang === 'en' ? 'Top players' : 'Топ игроков' }}</strong><button class="spin-button" :disabled="spinRolling || !sorted.length" @click="spinSMLT"><span aria-hidden="true">🎲</span>{{ spinRolling ? (lang === 'en' ? 'Rolling…' : 'Крутим…') : (lang === 'en' ? 'Spin SMLT' : 'Крутить SMLT') }}</button></div>
            <div class="filters">
              <label class="search-box"><span>⌕</span><input v-model="search" type="search" autocomplete="off" :placeholder="lang === 'en' ? 'Search player…' : 'Поиск игрока…'"></label>
              <label class="select-box"><span class="sr-only">{{ lang === 'en' ? 'Country' : 'Страна' }}</span><select v-model="country"><option value="ALL">{{ lang === 'en' ? 'All countries' : 'Все страны' }}</option><option v-for="code in countries" :key="code" :value="code">{{ countryMeta(code, lang).name }}</option></select></label>
            </div>
          </div>

          <div class="leaderboard-stats">
            <div class="stat"><strong>{{ players.length }}</strong><small>{{ lang === 'en' ? 'Total players' : 'Всего игроков' }}</small></div>
            <div class="stat"><strong>{{ average.toFixed(2) }}</strong><small>{{ lang === 'en' ? 'Average points' : 'Средние очки' }}</small></div>
            <div class="stat accent"><strong>{{ total.toFixed(2) }}</strong><small>{{ lang === 'en' ? 'Total points' : 'Общее количество очков' }}</small></div>
          </div>

          <div v-if="loading" class="large-state"><span class="loader"></span><p>{{ lang === 'en' ? 'Loading leaderboard…' : 'Загружаем рейтинг…' }}</p></div>
          <div v-else-if="error" class="large-state error-state"><p>{{ error }}</p><button class="primary-button" @click="$emit('retry')">{{ lang === 'en' ? 'Try again' : 'Повторить' }}</button></div>
          <template v-else>
            <div class="table-header"><span>{{ lang === 'en' ? 'Top' : 'Топ' }}</span><span>{{ lang === 'en' ? 'Country' : 'Страна' }}</span><span>{{ lang === 'en' ? 'Player' : 'Игрок' }}</span><span>{{ lang === 'en' ? 'Points' : 'Очки' }}</span><span>{{ lang === 'en' ? 'Hardest' : 'Хардест' }}</span><span>{{ lang === 'en' ? 'Global rank' : 'Мировой рейтинг' }}</span><span v-if="isAdmin"></span></div>
            <div class="player-list">
              <article v-for="player in visible" :key="player.id" :class="['player-row', {'top-three': player.rank <= 3}]">
                <span class="rank">{{ player.rank }}</span>
                <img v-if="countryMeta(player.country, lang).flagSrc" class="flag" :src="countryMeta(player.country, lang).flagSrc" :alt="countryMeta(player.country, lang).name" :title="countryMeta(player.country, lang).name"><span v-else class="flag" :title="countryMeta(player.country, lang).name">{{ countryMeta(player.country, lang).flag }}</span>
                <div class="player-cell"><button class="player-name-button" @click.stop="selected = player">{{ player.name }}</button><small class="mobile-demon">{{ player.demon }}</small></div>
                <strong class="points">{{ Number(player.points).toFixed(2) }}</strong>
                <span class="demon">{{ player.demon || '—' }}</span>
                <span class="global-rank">#{{ player.globalRank || '—' }}</span>
                <div v-if="isAdmin" class="row-actions"><button title="Edit" @click="$emit('edit', player)">✎</button><button title="Delete" @click="$emit('delete', player)">×</button></div>
              </article>
              <p v-if="!visible.length" class="no-results">{{ lang === 'en' ? 'No players found' : 'Игроки не найдены' }}</p>
            </div>
          </template>
        </section>

        <RecentChanges :changes="changes" :players="players" :lang="lang" />
      </div>

      <aside class="panel countries-sidebar">
        <div class="panel-title"><h2>{{ lang === 'en' ? 'Players by country' : 'Игроки по странам' }}</h2></div>
        <div class="country-list"><button v-for="([code, count]) in countryStats" :key="code" @click="country = code"><img v-if="countryMeta(code, lang).flagSrc" class="flag" :src="countryMeta(code, lang).flagSrc" :alt="countryMeta(code, lang).name"><span v-else>{{ countryMeta(code, lang).flag }}</span><span>{{ countryMeta(code, lang).name }}</span><strong>{{ count }}</strong></button></div>
      </aside>
    </div>

    <PlayerDetailsView v-if="selected" :player="selected" :players="sorted" :lang="lang" @close="selected = null" />
    <div v-if="spinPlayer" class="spin-backdrop" @click.self="spinPlayer = null">
      <section class="spin-card" role="dialog" aria-modal="true" :aria-label="lang === 'en' ? 'SMLT random player' : 'Случайный игрок SMLT'">
        <button class="modal-close" :aria-label="lang === 'en' ? 'Close' : 'Закрыть'" @click="spinPlayer = null">×</button>
        <p class="eyebrow">{{ lang === 'en' ? 'SMLT RANDOMIZER' : 'СЛУЧАЙНЫЙ ИГРОК SMLT' }}</p>
        <div class="spin-die" aria-hidden="true">{{ spinRolling ? '🎲' : '✦' }}</div>
        <h2>{{ spinPlayer.name }}</h2>
        <p class="spin-country"><img v-if="countryMeta(spinPlayer.country, lang).flagSrc" class="flag" :src="countryMeta(spinPlayer.country, lang).flagSrc" alt=""><span v-else>{{ countryMeta(spinPlayer.country, lang).flag }}</span>{{ countryMeta(spinPlayer.country, lang).name }}</p>
        <p class="spin-points">{{ Number(spinPlayer.points).toFixed(2) }} {{ lang === 'en' ? 'points' : 'очков' }} · #{{ spinPlayer.rank }}</p>
        <button class="primary-button" :disabled="spinRolling" @click="spinSMLT">{{ spinRolling ? (lang === 'en' ? 'Rolling…' : 'Крутим…') : (lang === 'en' ? 'Again' : 'Ещё раз') }}</button>
      </section>
    </div>
  </div>
</template>
