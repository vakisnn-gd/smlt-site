<script setup>
import { computed, ref } from 'vue'
import { countryMeta } from '../data'

const props = defineProps({
  players: {type: Array, default: () => []},
  lang: String,
  themeMode: String,
  roughMode: Boolean,
  canInstall: Boolean,
})
const emit = defineEmits(['theme', 'style', 'install', 'open-player'])

const sorted = computed(() => [...props.players].sort((a, b) => Number(b.points) - Number(a.points) || a.name.localeCompare(b.name)))
const firstId = ref('')
const secondId = ref('')
const spinPlayer = ref(null)
const spinRolling = ref(false)

const firstPlayer = computed(() => sorted.value.find(player => String(player.id) === String(firstId.value)) || null)
const secondPlayer = computed(() => sorted.value.find(player => String(player.id) === String(secondId.value)) || null)
const canCompare = computed(() => firstPlayer.value && secondPlayer.value && firstPlayer.value.id !== secondPlayer.value.id)

function selectPlayer(player, slot) {
  if (slot === 'first') firstId.value = String(player.id)
  else secondId.value = String(player.id)
}

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

function clearComparison() {
  firstId.value = ''
  secondId.value = ''
}
</script>

<template>
  <div class="page tools-page">
    <header class="page-heading tools-heading">
      <p class="eyebrow">{{ lang === 'en' ? 'SMLT TOOLBOX' : 'ИНСТРУМЕНТЫ SMLT' }}</p>
      <h1>{{ lang === 'en' ? 'Tools' : 'Инструменты' }}</h1>
      <span>{{ lang === 'en' ? 'Compare players, roll the leaderboard and tune the site.' : 'Сравнивайте игроков, крутите рейтинг и настраивайте сайт.' }}</span>
    </header>

    <section class="tool-card comparison-tool">
      <div class="tool-card-heading">
        <div><p class="eyebrow">{{ lang === 'en' ? 'PLAYER COMPARISON' : 'СРАВНЕНИЕ ИГРОКОВ' }}</p><h2>{{ lang === 'en' ? 'Who is stronger?' : 'Кто сильнее?' }}</h2></div>
        <span class="tool-badge">{{ lang === 'en' ? '2 players' : '2 игрока' }}</span>
      </div>
      <div class="compare-pickers">
        <label><span>{{ lang === 'en' ? 'First player' : 'Первый игрок' }}</span><select v-model="firstId"><option value="">{{ lang === 'en' ? 'Choose a player' : 'Выберите игрока' }}</option><option v-for="player in sorted" :key="`first-${player.id}`" :value="String(player.id)">{{ player.name }} · #{{ player.rank }}</option></select></label>
        <div class="compare-vs" aria-hidden="true">×</div>
        <label><span>{{ lang === 'en' ? 'Second player' : 'Второй игрок' }}</span><select v-model="secondId"><option value="">{{ lang === 'en' ? 'Choose a player' : 'Выберите игрока' }}</option><option v-for="player in sorted" :key="`second-${player.id}`" :value="String(player.id)">{{ player.name }} · #{{ player.rank }}</option></select></label>
      </div>
      <p v-if="firstPlayer && secondPlayer && firstPlayer.id === secondPlayer.id" class="tool-error">{{ lang === 'en' ? 'Choose two different players.' : 'Выберите двух разных игроков.' }}</p>
      <div v-if="canCompare" class="comparison-result">
        <article v-for="player in [firstPlayer, secondPlayer]" :key="player.id" class="compare-card"><button class="compare-player-name" type="button" @click="emit('open-player', player.gdlId || player.id)">{{ player.name }}</button><span>{{ countryMeta(player.country, lang).name }}</span><dl><div><dt>{{ lang === 'en' ? 'SMLT rank' : 'Место SMLT' }}</dt><dd>#{{ player.rank }}</dd></div><div><dt>{{ lang === 'en' ? 'Points' : 'Очки' }}</dt><dd>{{ Number(player.points).toFixed(2) }}</dd></div><div><dt>{{ lang === 'en' ? 'World rank' : 'Мировой ранг' }}</dt><dd>#{{ player.globalRank || '—' }}</dd></div><div><dt>{{ lang === 'en' ? 'Hardest demon' : 'Хардест' }}</dt><dd>{{ player.demon || '—' }}</dd></div></dl></article>
        <p class="compare-result-line">{{ lang === 'en' ? 'Difference in points' : 'Разница в очках' }}: <strong>{{ Math.abs(Number(firstPlayer.points) - Number(secondPlayer.points)).toFixed(2) }}</strong> · {{ lang === 'en' ? 'rank places' : 'мест по SMLT' }}: <strong>{{ Math.abs(Number(firstPlayer.rank) - Number(secondPlayer.rank)) }}</strong></p>
      </div>
      <button v-if="firstId || secondId" class="secondary-button tool-clear" type="button" @click="clearComparison">{{ lang === 'en' ? 'Clear comparison' : 'Очистить сравнение' }}</button>
    </section>

    <div class="tools-grid">
      <section class="tool-card">
        <div class="tool-card-heading"><div><p class="eyebrow">{{ lang === 'en' ? 'RANDOMIZER' : 'СЛУЧАЙНЫЙ ИГРОК' }}</p><h2>{{ lang === 'en' ? 'Spin SMLT' : 'Крутить SMLT' }}</h2></div><span class="tool-icon" aria-hidden="true">🎲</span></div>
        <p class="tool-description">{{ lang === 'en' ? 'Pick a random player from the current leaderboard.' : 'Выберите случайного игрока из текущего рейтинга.' }}</p>
        <button class="primary-button" :disabled="spinRolling || !sorted.length" @click="spinSMLT">{{ spinRolling ? (lang === 'en' ? 'Rolling…' : 'Крутим…') : (lang === 'en' ? 'Spin the leaderboard' : 'Крутить рейтинг') }}</button>
        <div v-if="spinPlayer" class="random-result"><strong>{{ spinPlayer.name }}</strong><span>#{{ spinPlayer.rank }} · {{ Number(spinPlayer.points).toFixed(2) }} {{ lang === 'en' ? 'points' : 'очков' }}</span><button type="button" @click="emit('open-player', spinPlayer.gdlId || spinPlayer.id)">{{ lang === 'en' ? 'Open profile ↗' : 'Открыть профиль ↗' }}</button></div>
      </section>

      <section class="tool-card">
        <div class="tool-card-heading"><div><p class="eyebrow">{{ lang === 'en' ? 'SITE SETTINGS' : 'НАСТРОЙКИ САЙТА' }}</p><h2>{{ lang === 'en' ? 'Make it yours' : 'Настроить под себя' }}</h2></div><span class="tool-icon" aria-hidden="true">✦</span></div>
        <div class="setting-row"><div><strong>{{ lang === 'en' ? 'Theme' : 'Тема' }}</strong><span>{{ themeMode === 'light' ? (lang === 'en' ? 'White' : 'Белая') : (lang === 'en' ? 'Dark' : 'Тёмная') }}</span></div><button class="secondary-button" type="button" @click="emit('theme')">{{ themeMode === 'light' ? '☾' : '☀' }} {{ themeMode === 'light' ? (lang === 'en' ? 'Dark' : 'Тёмная') : (lang === 'en' ? 'White' : 'Белая') }}</button></div>
        <div class="setting-row"><div><strong>{{ lang === 'en' ? 'Style' : 'Стиль' }}</strong><span>{{ roughMode ? (lang === 'en' ? 'Playful' : 'Рофл') : (lang === 'en' ? 'Clean' : 'Обычный') }}</span></div><button class="secondary-button" type="button" @click="emit('style')">{{ roughMode ? '✎' : '✦' }} {{ roughMode ? (lang === 'en' ? 'Clean' : 'Обычный') : (lang === 'en' ? 'Playful' : 'Рофл') }}</button></div>
        <div class="setting-row install-row"><div><strong>{{ lang === 'en' ? 'App' : 'Приложение' }}</strong><span>{{ lang === 'en' ? 'Install SMLT on your device' : 'Установить SMLT на устройство' }}</span></div><button v-if="canInstall" class="primary-button" type="button" @click="emit('install')">⇩ {{ lang === 'en' ? 'Install' : 'Установить' }}</button><span v-else class="install-hint">{{ lang === 'en' ? 'Use the browser menu' : 'Через меню браузера' }}</span></div>
      </section>
    </div>
  </div>
</template>
