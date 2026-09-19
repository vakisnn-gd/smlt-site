<script setup>
import { onMounted, ref } from 'vue'
import { api } from '../api'

const props = defineProps({lang: String})
const emit = defineEmits(['close', 'authenticated'])
const captcha = ref(null)
const password = ref('')
const answer = ref('')
const error = ref('')
const loading = ref(false)

async function refreshCaptcha() {
  error.value = ''
  answer.value = ''
  try { captcha.value = await api.captcha() } catch (err) { error.value = err.message }
}

async function submit() {
  if (!captcha.value || !password.value || !answer.value) return
  loading.value = true
  error.value = ''
  try {
    await api.login({password: password.value, captchaId: captcha.value.captchaId, captchaAnswer: answer.value})
    emit('authenticated')
  } catch (err) {
    error.value = err.message
    await refreshCaptcha()
  } finally { loading.value = false }
}

onMounted(refreshCaptcha)
</script>

<template>
  <div class="modal-backdrop" @mousedown.self="$emit('close')">
    <section class="modal small-modal" role="dialog" aria-modal="true">
      <button class="modal-close" @click="$emit('close')">×</button>
      <p class="eyebrow">SMLT CONTROL</p><h2>{{ lang === 'en' ? 'Admin access' : 'Вход администратора' }}</h2><p class="modal-subtitle">{{ lang === 'en' ? 'Enter the server password and solve the captcha.' : 'Введите пароль сервера и решите капчу.' }}</p>
      <form @submit.prevent="submit">
        <label>{{ lang === 'en' ? 'Password' : 'Пароль' }}<input v-model="password" type="password" autocomplete="current-password" required></label>
        <label>{{ lang === 'en' ? 'Captcha' : 'Капча' }}
          <div class="captcha-row"><div class="captcha-frame"><img v-if="captcha" :src="captcha.imageUrl" alt="Captcha"><span v-else class="mini-loader"></span></div><button type="button" class="refresh-button" @click="refreshCaptcha">↻</button></div>
        </label>
        <label>{{ lang === 'en' ? 'Answer' : 'Ответ' }}<input v-model="answer" autocomplete="off" required></label>
        <p v-if="error" class="form-error">{{ error }}</p>
        <button class="primary-button full-button" :disabled="loading">{{ loading ? '…' : (lang === 'en' ? 'Sign in' : 'Войти') }}</button>
      </form>
    </section>
  </div>
</template>
