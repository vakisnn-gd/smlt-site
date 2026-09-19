<script setup>
import { computed, onBeforeUnmount, onMounted } from 'vue'
import { countryMeta } from '../data'

const props = defineProps({player: Object, lang: String, players: Array})
const emit = defineEmits(['close'])

const sorted = computed(() => [...(props.players || [])].sort((a, b) => Number(b.points) - Number(a.points) || a.name.localeCompare(b.name)))
const total = computed(() => sorted.value.length)
const index = computed(() => sorted.value.findIndex(p => p.id === props.player.id))
const above = computed(() => index.value > 0 ? sorted.value[index.value - 1] : null)
const below = computed(() => (index.value >= 0 && index.value < total.value - 1) ? sorted.value[index.value + 1] : null)
const countryRank = computed(() => {
  const same = sorted.value.filter(p => p.country === props.player.country)
  const pos = same.findIndex(p => p.id === props.player.id)
  return {pos: pos + 1, total: same.length}
})

function onKeydown(event) {
  if (event.key === 'Escape') emit('close')
}
onMounted(() => window.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <div class="modal-backdrop" @mousedown.self="$emit('close')">
    <section class="modal player-details-modal" role="dialog" aria-modal="true">
      <button class="modal-close" @click="$emit('close')">×</button>
      <div class="pdetails-hero">
        <span class="pdetails-flag">{{ countryMeta(player.country, lang).flag }}</span>
        <div>
          <p class="eyebrow">SMLT PLAYER</p>
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
        <span class="pdetails-chip">{{ lang === 'en' ? 'Player' : 'Игрок' }} {{ countryRank.pos }} / {{ countryRank.total }} {{ lang === 'en' ? 'of' : 'из' }} {{ countryMeta(player.country, lang).name }}</span>
        <a v-if="player.gdlId" class="pdetails-chip pdetails-link" :href="`https://demonlist.org/profile/${player.gdlId}`" target="_blank" rel="noopener">Demonlist ↗</a>
      </div>
      <div v-if="above || below" class="pdetails-neighbors">
        <button v-if="above" class="pdetails-neighbor" @click="$emit('close'); $emit('open', above)"><small>{{ lang === 'en' ? 'Above' : 'Выше' }} · #{{ above.rank }}</small><span>{{ above.name }}</span></button>
        <button v-else class="pdetails-neighbor placeholder"><small>{{ lang === 'en' ? 'Above' : 'Выше' }}</small><span>—</span></button>
        <button v-if="below" class="pdetails-neighbor" @click="$emit('close'); $emit('open', below)"><small>{{ lang === 'en' ? 'Below' : 'Ниже' }} · #{{ below.rank }}</small><span>{{ below.name }}</span></button>
        <button v-else class="pdetails-neighbor placeholder"><small>{{ lang === 'en' ? 'Below' : 'Ниже' }}</small><span>—</span></button>
      </div>
      <a v-if="player.gdlId" :href="`https://demonlist.org/profile/${player.gdlId}`" class="primary-button pdetails-open" target="_blank" rel="noopener">{{ lang === 'en' ? 'Open on Demonlist' : 'Открыть на Demonlist' }}</a>
    </section>
  </div>
</template>