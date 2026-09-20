<script setup>
import { computed } from 'vue'

const props = defineProps({changes: {type: Array, default: () => []}, players: {type: Array, default: () => []}, lang: String})

const playerMap = computed(() => new Map(props.players.map(player => [player.name.toLowerCase(), player])))

function playerUrl(name) {
  const player = playerMap.value.get(String(name).toLowerCase())
  return player?.gdlId ? `https://demonlist.org/profile/${player.gdlId}` : ''
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
    <p v-if="!changes.length" class="empty-state">{{ lang === 'en' ? 'Position changes will appear after the next leaderboard update.' : 'Изменения позиций появятся после следующего обновления рейтинга.' }}</p>
    <div v-else class="change-scroll">
      <template v-for="group in grouped" :key="group.label">
        <div class="date-divider"><span>{{ group.label }}</span></div>
        <article v-for="change in group.changes" :key="change.id" class="change-row">
          <span class="change-arrow">↑</span>
          <div>
            <small class="change-date">{{ dateTime(change.createdAt) }}</small>
            <p>
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
