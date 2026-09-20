<script setup>
defineProps({lang: String, isAdmin: Boolean})
defineEmits(['leaderboard', 'events', 'login', 'logout'])
</script>

<template>
  <div class="page community-page">
    <section class="community-intro">
      <h1>{{ lang === 'en' ? 'What is SMLT?' : 'Что такое SMLT?' }}</h1>
      <p>{{ lang === 'en' ? 'SMLT is a Discord server where we hold events: collaborations, tournaments, level part runs, and even Minecraft games.' : 'SMLT — это дискорд-сервер, где происходят разные ивенты: коллабы, турниры, прохождения уровней по частям и даже игры в Майнкрафте.' }}</p>
      <p>{{ lang === 'en' ? 'Every member of the server is a member of SMLT, regardless of their skill. The idea behind our server is ' : 'Все участники сервера — участники SMLT, независимо от их скилла. Идея нашего сервера — ' }}<strong>{{ lang === 'en' ? 'fun.' : 'удовольствие.' }}</strong></p>
    </section>
    <div class="community-links">
      <button @click="$emit('leaderboard')"><strong>{{ lang === 'en' ? 'Leaderboard' : 'Лидерборд' }}</strong><span>{{ lang === 'en' ? 'The best SMLT players' : 'Лучшие игроки SMLT' }}</span></button>
      <button @click="$emit('events')"><strong>{{ lang === 'en' ? 'SMLT projects' : 'Проекты SMLT' }}</strong><span>{{ lang === 'en' ? 'Collabs and community events' : 'Коллабы и ивенты сообщества' }}</span></button>
    </div>
    <section class="community-info">
      <h2>{{ lang === 'en' ? 'Information' : 'Информация' }}</h2>
      <div class="community-contact"><h3>{{ lang === 'en' ? 'SMLT host' : 'Хост SMLT' }}</h3><p>Discord: <a href="https://discord.com/users/.samoletik" target="_blank" rel="noopener">@.samoletik</a></p><p>Telegram: <a href="https://t.me/samoletik" target="_blank" rel="noopener">@samoletik</a></p><p>GitHub: <a href="https://github.com/samoletix" target="_blank" rel="noopener">@samoletix</a></p></div>
      <div class="community-contact"><h3>{{ lang === 'en' ? 'Site admin' : 'Админ сайта' }}</h3><p>Discord: <a href="https://discord.com/users/782665885441458208" target="_blank" rel="noopener">@gmdprocessor</a></p><p>Telegram: <a href="https://t.me/gmdprocessor" target="_blank" rel="noopener">@gmdprocessor</a></p><p>GitHub: <a href="https://github.com/vakisnn-gd" target="_blank" rel="noopener">@vakisnn-gd</a></p></div>
      <div class="community-contact"><h3>{{ lang === 'en' ? 'Former admin' : 'Бывший админ' }}</h3><p>Discord: <a href="https://discord.com/users/rimix.98" target="_blank" rel="noopener">@rimix.98</a></p><p>Telegram: <a href="https://t.me/Rimix980" target="_blank" rel="noopener">@Rimix980</a></p><p>GitHub: <a href="https://github.com/rimix98" target="_blank" rel="noopener">@rimix98</a></p></div>
      <div class="community-admin-access">
        <span>{{ lang === 'en' ? 'Site administration' : 'Управление сайтом' }}</span>
        <button v-if="!isAdmin" class="secondary-button" @click="$emit('login')">{{ lang === 'en' ? 'Admin sign in' : 'Войти в админку' }}</button>
        <button v-else class="secondary-button" @click="$emit('logout')">{{ lang === 'en' ? 'Sign out' : 'Выйти из админки' }}</button>
      </div>
    </section>
  </div>
</template>
