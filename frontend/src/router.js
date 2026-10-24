import {createRouter, createWebHistory} from 'vue-router'

const EmptyRoute = {render: () => null}

export default createRouter({
  history: createWebHistory(),
  routes: [
    {path: '/', name: 'leaderboard', component: EmptyRoute, meta: {view: 'leaderboard', lang: 'ru'}},
    {path: '/events', name: 'events', component: EmptyRoute, meta: {view: 'events', lang: 'ru'}},
    {path: '/chats', name: 'chats', component: EmptyRoute, meta: {view: 'chats', lang: 'ru'}},
    {path: '/tools', name: 'tools', component: EmptyRoute, meta: {view: 'tools', lang: 'ru'}},
    {path: '/about', name: 'about', component: EmptyRoute, meta: {view: 'about', lang: 'ru'}},
    {path: '/player/:id(\\d+)', name: 'player', component: EmptyRoute, meta: {view: 'leaderboard', lang: 'ru'}},
    {path: '/en', name: 'en-leaderboard', component: EmptyRoute, meta: {view: 'leaderboard', lang: 'en'}},
    {path: '/en/events', name: 'en-events', component: EmptyRoute, meta: {view: 'events', lang: 'en'}},
    {path: '/en/chats', name: 'en-chats', component: EmptyRoute, meta: {view: 'chats', lang: 'en'}},
    {path: '/en/tools', name: 'en-tools', component: EmptyRoute, meta: {view: 'tools', lang: 'en'}},
    {path: '/en/about', name: 'en-about', component: EmptyRoute, meta: {view: 'about', lang: 'en'}},
    {path: '/en/player/:id(\\d+)', name: 'en-player', component: EmptyRoute, meta: {view: 'leaderboard', lang: 'en'}},
  ],
  scrollBehavior(to, from) {
    if (String(to.name).endsWith('player') || String(from.name).endsWith('player')) return false
    return {top: 0, behavior: 'smooth'}
  },
})
