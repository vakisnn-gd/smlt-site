<script setup>
import { onBeforeUnmount, ref } from 'vue'

defineProps({view: String, lang: String, isAdmin: Boolean, roughMode: Boolean})
const emit = defineEmits(['navigate', 'language', 'style', 'add', 'logout', 'easter-egg'])
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
      <button class="brand" @click="brandClick"><img class="brand-favicon" src="/favicon2.ico" alt=""><strong>SMLT</strong></button>
      <nav :class="['primary-nav', {open: menuOpen}]">
        <button :class="{active: view === 'leaderboard'}" @click="go('leaderboard')">{{ lang === 'en' ? 'Leaderboard' : 'Рейтинг' }}</button>
        <button :class="{active: view === 'events'}" @click="go('events')">{{ lang === 'en' ? 'Events' : 'Ивенты' }}</button>
        <button :class="{active: view === 'about'}" @click="go('about')">{{ lang === 'en' ? 'About' : 'О SMLT' }}</button>
      </nav>
      <div class="nav-actions">
        <button class="style-toggle" :class="{active: roughMode}" :aria-pressed="roughMode" :title="roughMode ? (lang === 'en' ? 'Switch to clean style' : 'Вернуть обычный стиль') : (lang === 'en' ? 'Turn on playful style' : 'Включить рофляный стиль')" @click="$emit('style')"><span aria-hidden="true">{{ roughMode ? '✎' : '✦' }}</span><span class="style-toggle-label">{{ roughMode ? (lang === 'en' ? 'Playful' : 'Рофл') : (lang === 'en' ? 'Clean' : 'Обычный') }}</span></button>
        <button class="lang-button" @click="$emit('language')">{{ lang === 'en' ? 'RU' : 'EN' }}</button>
        <template v-if="isAdmin">
          <button class="icon-button add-button" :title="lang === 'en' ? 'Add player' : 'Добавить игрока'" @click="$emit('add')">＋</button>
          <button class="icon-button" :title="lang === 'en' ? 'Log out' : 'Выйти'" @click="$emit('logout')">↪</button>
        </template>
      </div>
    </div>
  </header>
</template>
