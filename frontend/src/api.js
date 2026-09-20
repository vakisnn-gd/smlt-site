async function request(path, options = {}) {
  const response = await fetch(path, {
    credentials: 'same-origin',
    ...options,
    headers: options.body ? {'Content-Type': 'application/json', ...(options.headers || {})} : options.headers,
  })
  const data = await response.json().catch(() => ({}))
  if (!response.ok || data.success === false) {
    const error = new Error(data.message || `HTTP ${response.status}`)
    error.status = response.status
    throw error
  }
  return data
}

export const api = {
  players: () => request(`/api/players?fresh=${Date.now()}`),
  changes: () => request(`/api/recent-changes?fresh=${Date.now()}`),
  events: () => request(`/api/events?fresh=${Date.now()}`),
  addEvent: event => request('/api/events', {method: 'POST', body: JSON.stringify(event)}),
  deleteEvent: id => request(`/api/events/${id}`, {method: 'DELETE'}),
  reorderEvents: ids => request('/api/events', {method: 'PUT', body: JSON.stringify({ids})}),
  session: () => request('/api/session'),
  captcha: () => request('/api/captcha'),
  login: payload => request('/api/auth', {method: 'POST', body: JSON.stringify(payload)}),
  logout: () => request('/api/logout', {method: 'POST'}),
  search: name => request(`/api/search-demonlist?name=${encodeURIComponent(name)}`),
  levels: (gdlId, name) => request(`/api/player-levels?id=${encodeURIComponent(gdlId || '')}&name=${encodeURIComponent(name || '')}`),
  addPlayer: player => request('/api/players', {method: 'POST', body: JSON.stringify(player)}),
  updatePlayer: (oldName, player) => request(`/api/players/${encodeURIComponent(oldName)}`, {method: 'PUT', body: JSON.stringify(player)}),
  deletePlayer: name => request(`/api/players/${encodeURIComponent(name)}`, {method: 'DELETE'}),
}
