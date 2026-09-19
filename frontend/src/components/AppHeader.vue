<script setup>
import { ref } from 'vue'

defineProps({view: String, lang: String, isAdmin: Boolean})
const emit = defineEmits(['navigate', 'language', 'login', 'add', 'logout'])
const menuOpen = ref(false)

function go(view) {
  emit('navigate', view)
  menuOpen.value = false
}
</script>

<template>
  <header class="site-header">
    <div class="nav-wrap">
      <button class="mobile-menu" :aria-expanded="menuOpen" :aria-label="lang === 'en' ? 'Open menu' : 'Открыть меню'" @click="menuOpen = !menuOpen"><span></span><span></span><span></span></button>
      <button class="brand" @click="go('leaderboard')"><img class="brand-favicon" src="/favicon2.ico" alt=""><strong>SMLT</strong></button>
      <nav :class="['primary-nav', {open: menuOpen}]">
        <button :class="{active: view === 'leaderboard'}" @click="go('leaderboard')">{{ lang === 'en' ? 'Leaderboard' : 'Рейтинг' }}</button>
        <button :class="{active: view === 'events'}" @click="go('events')">{{ lang === 'en' ? 'Events' : 'Ивенты' }}</button>
        <button :class="{active: view === 'about'}" @click="go('about')">{{ lang === 'en' ? 'About' : 'О SMLT' }}</button>
      </nav>
      <div class="nav-actions">
        <a class="social-link" href="https://discord.gg/VK56W7ZzdA" target="_blank" rel="noopener">Discord</a>
        <button class="lang-button" @click="$emit('language')">{{ lang === 'en' ? 'RU' : 'EN' }}</button>
        <template v-if="isAdmin">
          <button class="icon-button add-button" :title="lang === 'en' ? 'Add player' : 'Добавить игрока'" @click="$emit('add')">＋</button>
          <button class="icon-button" :title="lang === 'en' ? 'Log out' : 'Выйти'" @click="$emit('logout')">↪</button>
        </template>
        <button v-else class="admin-button" @click="$emit('login')">{{ lang === 'en' ? 'Admin' : 'Админ' }}</button>
      </div>
    </div>
  </header>
</template>
