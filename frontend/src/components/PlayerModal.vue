<script setup>
import { reactive, ref } from 'vue'
import { api } from '../api'
import { countries } from '../data'

const props = defineProps({lang: String, player: Object})
const emit = defineEmits(['close', 'saved', 'unauthorized'])
const form = reactive(props.player ? {...props.player} : {name: '', country: 'OTHER', points: 0, demon: '—', globalRank: 0})
const saving = ref(false)
const error = ref('')

async function save() {
  saving.value = true
  error.value = ''
  try {
    if (props.player) {
      const payload = {name: form.name.trim(), country: form.country, points: Number(form.points), demon: form.demon.trim() || '—', globalRank: Number(form.globalRank)}
      await api.updatePlayer(props.player.name, payload)
    } else {
      const found = await api.search(form.name.trim())
      await api.addPlayer({name: found.name, country: String(found.country || 'OTHER').toUpperCase(), points: found.points, demon: found.demon || '—', globalRank: found.globalRank})
    }
    emit('saved')
  } catch (err) {
    error.value = err.message
    if (err.status === 401) emit('unauthorized')
  } finally { saving.value = false }
}
</script>

<template>
  <div class="modal-backdrop" @mousedown.self="$emit('close')">
    <section class="modal player-modal" role="dialog" aria-modal="true">
      <button class="modal-close" @click="$emit('close')">×</button>
      <p class="eyebrow">{{ player ? 'EDIT PLAYER' : 'DEMONLIST' }}</p><h2>{{ player ? (lang === 'en' ? 'Edit player' : 'Изменить игрока') : (lang === 'en' ? 'Add player' : 'Добавить игрока') }}</h2>
      <form @submit.prevent="save">
        <p v-if="!player" class="modal-subtitle">{{ lang === 'en' ? 'Enter a Demonlist nickname. The rest will be filled in automatically.' : 'Введите ник из Demonlist — остальное сайт заполнит сам.' }}</p>
        <label>{{ lang === 'en' ? 'Nickname' : 'Никнейм' }}<input v-model="form.name" maxlength="30" required autofocus></label>
        <template v-if="player"><div class="form-columns"><label>{{ lang === 'en' ? 'Country' : 'Страна' }}<select v-model="form.country"><option v-for="(_, code) in countries" :key="code" :value="code">{{ code }}</option></select></label><label>{{ lang === 'en' ? 'World rank' : 'Место в мире' }}<input v-model.number="form.globalRank" type="number" min="0" required></label></div>
        <label>{{ lang === 'en' ? 'Points' : 'Очки' }}<input v-model.number="form.points" type="number" min="0" step="0.01" required></label>
        <label>{{ lang === 'en' ? 'Hardest demon' : 'Сложнейший демон' }}<input v-model="form.demon" maxlength="50" required></label></template>
        <p v-if="error" class="form-error">{{ error }}</p>
        <div class="modal-actions"><button type="button" class="secondary-button" @click="$emit('close')">{{ lang === 'en' ? 'Cancel' : 'Отмена' }}</button><button class="primary-button" :disabled="saving">{{ saving ? '…' : (player ? (lang === 'en' ? 'Save' : 'Сохранить') : (lang === 'en' ? 'Add' : 'Добавить')) }}</button></div>
      </form>
    </section>
  </div>
</template>
