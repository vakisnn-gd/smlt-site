<script setup>
import { onBeforeUnmount, ref } from 'vue'

defineProps({view: String, lang: String, isAdmin: Boolean, roughMode: Boolean})
const emit = defineEmits(['navigate', 'language', 'add', 'logout', 'easter-egg'])
const menuOpen = ref(false)
const brandClicks = ref(0)
let brandClickTimer

function go(view) {
  emit('navigate', view)
  menuOpen.value = false
}

function brandClick() {
  go('leaderboard')
  brandClicks.value += 1
  clearTimeout(brandClickTimer)
  if (brandClicks.value >= 5) {
    brandClicks.value = 0
    emit('easter-egg')
    return
  }
  brandClickTimer = setTimeout(() => { brandClicks.value = 0 }, 1400)
}

onBeforeUnmount(() => clearTimeout(brandClickTimer))
</script>

<template>
  <header class="site-header">
    <div class="nav-wrap">
      <button class="mobile-menu" :aria-expanded="menuOpen" :aria-label="lang === 'en' ? 'Open menu' : 'Открыть меню'" @click="menuOpen = !menuOpen"><span></span><span></span><span></span></button>
      <button class="brand" @click="brandClick"><img class="brand-favicon" :src="roughMode ? '/avatars/smlt-rough.webp' : '/favicon2.ico'" alt=""><strong>SMLT</strong></button>
      <nav :class="['primary-nav', {open: menuOpen}]" :aria-label="lang === 'en' ? 'Main navigation' : 'Основная навигация'" @keydown.esc="menuOpen = false">
        <button :class="{active: view === 'leaderboard'}" :aria-current="view === 'leaderboard' ? 'page' : undefined" @click="go('leaderboard')">{{ lang === 'en' ? 'Leaderboard' : 'Рейтинг' }}</button>
        <button :class="{active: view === 'events'}" :aria-current="view === 'events' ? 'page' : undefined" @click="go('events')">{{ lang === 'en' ? 'Events' : 'Ивенты' }}</button>
        <button :class="{active: view === 'chats'}" :aria-current="view === 'chats' ? 'page' : undefined" @click="go('chats')">{{ lang === 'en' ? 'Chats' : 'Чаты' }}</button>
        <button :class="{active: view === 'tools'}" :aria-current="view === 'tools' ? 'page' : undefined" @click="go('tools')">{{ lang === 'en' ? 'Tools' : 'Инструменты' }}</button>
        <button :class="{active: view === 'about'}" :aria-current="view === 'about' ? 'page' : undefined" @click="go('about')">{{ lang === 'en' ? 'About' : 'О SMLT' }}</button>
      </nav>
      <div class="nav-actions">
        <button class="lang-button" @click="$emit('language')">{{ lang === 'en' ? 'RU' : 'EN' }}</button>
        <template v-if="isAdmin">
          <button class="icon-button add-button" :title="lang === 'en' ? 'Add player' : 'Добавить игрока'" @click="$emit('add')">＋</button>
          <button class="icon-button" :title="lang === 'en' ? 'Log out' : 'Выйти'" @click="$emit('logout')">↪</button>
        </template>
      </div>
    </div>
  </header>
</template>
