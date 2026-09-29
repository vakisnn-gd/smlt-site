<script setup>
import { computed } from 'vue'
import { flagSrc } from '../data'

const props = defineProps({changes: {type: Array, default: () => []}, players: {type: Array, default: () => []}, lang: String})

const playerMap = computed(() => new Map(props.players.map(player => [player.name.toLowerCase(), player])))

function playerUrl(name) {
  const player = playerMap.value.get(String(name).toLowerCase())
  return player?.gdlId ? `https://demonlist.org/profile/${player.gdlId}` : ''
}

function eventCountry(change) {
  return change.country || playerMap.value.get(String(change.playerName || '').toLowerCase())?.country || 'OTHER'
}

function dateLabel(value) {
  const date = new Date(value)
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const target = new Date(date.getFullYear(), date.getMonth(), date.getDate())
  const days = Math.round((today - target) / 86400000)
  if (days === 0) return props.lang === 'en' ? 'TODAY' : 'СЕГОДНЯ'
  if (days === 1) return props.lang === 'en' ? 'YESTERDAY' : 'ВЧЕРА'
  return new Intl.DateTimeFormat(props.lang === 'en' ? 'en-US' : 'ru-RU', {day: 'numeric', month: 'long', year: 'numeric'}).format(date).toUpperCase()
}

function dateTime(value) {
  const date = new Date(value)
  const locale = props.lang === 'en' ? 'en-US' : 'ru-RU'
  const day = new Intl.DateTimeFormat(locale, {day: 'numeric', month: 'short', year: 'numeric'}).format(date)
  const time = new Intl.DateTimeFormat(locale, {hour: '2-digit', minute: '2-digit'}).format(date)
  return `${day}, ${time}`
}

const grouped = computed(() => {
  const groups = []
  for (const change of props.changes.slice(0, 20)) {
    const label = dateLabel(change.createdAt)
    let group = groups.find(item => item.label === label)
    if (!group) { group = {label, changes: []}; groups.push(group) }
    group.changes.push(change)
  }
  return groups
})
</script>

<template>
  <section class="panel recent-panel">
    <div class="panel-title"><span class="title-icon">↻</span><h2>{{ lang === 'en' ? 'Recent changes' : 'Недавние изменения' }}</h2></div>
    <p v-if="!changes.length" class="empty-state">{{ lang === 'en' ? 'Player additions, removals and position changes will appear here.' : 'Здесь появятся добавления, удаления игроков и изменения позиций.' }}</p>
    <div v-else class="change-scroll">
      <template v-for="group in grouped" :key="group.label">
        <div class="date-divider"><span>{{ group.label }}</span></div>
        <article v-for="change in group.changes" :key="`${change.eventType || 'rank'}-${change.id}`" class="change-row">
          <span class="change-arrow">{{ change.eventType === 'added' ? '＋' : change.eventType === 'removed' ? '−' : change.eventType === 'demon' ? '◆' : '↑' }}</span>
          <div>
            <small class="change-date">{{ dateTime(change.createdAt) }}</small>
            <p v-if="change.eventType">
              <img class="country-flag" :src="flagSrc(eventCountry(change), 40)" :alt="eventCountry(change)" :title="eventCountry(change)">
              <strong>{{ change.playerName }}</strong>
              {{ change.eventType === 'added' ? (lang === 'en' ? ' was added to the leaderboard' : ' добавлен в список') : change.eventType === 'removed' ? (lang === 'en' ? ' was removed from the leaderboard' : ' удалён из списка') : (lang === 'en' ? ' completed a new demon: ' : ' прошёл новый демон: ') }}<strong v-if="change.eventType === 'demon'">{{ change.detail }}</strong>
            </p>
            <p v-else>
              <img class="country-flag" :src="flagSrc(eventCountry(change), 40)" :alt="eventCountry(change)" :title="eventCountry(change)">
              <a v-if="playerUrl(change.playerName)" :href="playerUrl(change.playerName)" target="_blank" rel="noopener">{{ change.playerName }}</a><strong v-else>{{ change.playerName }}</strong>
              {{ lang === 'en' ? ` moved from #${change.oldRank} to #${change.newRank}` : ` поднялся с #${change.oldRank} на #${change.newRank}` }}
              <template v-if="change.passedPlayers?.length">{{ lang === 'en' ? ', passing ' : ', обойдя ' }}<template v-for="(name, index) in change.passedPlayers.slice(0, 4)" :key="name"><span v-if="index">{{ index === Math.min(change.passedPlayers.length, 4) - 1 ? (lang === 'en' ? ' and ' : ' и ') : ', ' }}</span><a v-if="playerUrl(name)" :href="playerUrl(name)" target="_blank" rel="noopener">{{ name }}</a><strong v-else>{{ name }}</strong></template></template>
            </p>
            <small v-if="change.abovePlayer && change.belowPlayer">{{ lang === 'en' ? 'Now between' : 'Теперь между' }} <span>{{ change.abovePlayer }}</span> {{ lang === 'en' ? 'and' : 'и' }} <span>{{ change.belowPlayer }}</span></small>
          </div>
        </article>
      </template>
    </div>
  </section>
</template>
