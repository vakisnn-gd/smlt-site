import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './styles.css'

const app = createApp(App)
app.use(router)
router.isReady().then(() => {
  const root = document.querySelector('#app')
  app.mount(root)
  root.hidden = false
})

if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/service-worker.js').catch(error => {
      console.warn('[pwa] service worker registration failed', error)
    })
  })
}
