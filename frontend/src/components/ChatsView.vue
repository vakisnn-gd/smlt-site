<script setup>
import { onMounted } from 'vue'

defineProps({lang: String})

onMounted(() => {
  window.location.replace('https://chat.smlt.lol/')
})
</script>

<template>
  <section class="chats-page" :aria-label="lang === 'en' ? 'SMLT chats' : 'Чаты SMLT'">
    <p>{{ lang === 'en' ? 'Opening SMLT chat…' : 'Открываем чат SMLT…' }}</p>
    <a href="https://chat.smlt.lol/">{{ lang === 'en' ? 'Open chat' : 'Перейти в чат' }}</a>
  </section>
</template>

<style scoped>
.chats-page {
  width: min(900px, 100%);
  margin: 0 auto;
  padding: 48px 22px;
  color: var(--text, #e8edf5);
  text-align: center;
}

.chats-page a {
  color: var(--accent, #63adff);
}
</style>
