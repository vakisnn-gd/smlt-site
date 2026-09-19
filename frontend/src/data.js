export const countries = {
  RU: ['🇷🇺', 'Россия', 'Russia'], UA: ['🇺🇦', 'Украина', 'Ukraine'], BY: ['🇧🇾', 'Беларусь', 'Belarus'],
  RS: ['🇷🇸', 'Сербия', 'Serbia'], AM: ['🇦🇲', 'Армения', 'Armenia'], BG: ['🇧🇬', 'Болгария', 'Bulgaria'],
  DE: ['🇩🇪', 'Германия', 'Germany'], KZ: ['🇰🇿', 'Казахстан', 'Kazakhstan'], US: ['🇺🇸', 'США', 'United States'],
  GB: ['🇬🇧', 'Великобритания', 'United Kingdom'], IT: ['🇮🇹', 'Италия', 'Italy'], TR: ['🇹🇷', 'Турция', 'Türkiye'],
  NL: ['🇳🇱', 'Нидерланды', 'Netherlands'], PL: ['🇵🇱', 'Польша', 'Poland'], FR: ['🇫🇷', 'Франция', 'France'],
  CA: ['🇨🇦', 'Канада', 'Canada'], BR: ['🇧🇷', 'Бразилия', 'Brazil'], OTHER: ['🌐', 'Другое', 'Other'],
}

export function countryMeta(code, lang = 'ru') {
  const entry = countries[String(code || '').toUpperCase()] || [String(code || '🌐'), String(code || 'Другое'), String(code || 'Other')]
  return {flag: entry[0], name: entry[lang === 'en' ? 2 : 1]}
}

export const eventVideos = [
  ['axsCwtl2Hro', 'SMLT beats Back on Track', 'beat'], ['58ItHCC8ZKI', 'SMLT beats Every End', 'beat'], ['8RduRoZmFHs', 'SMLT beats Ne Rofl Collab', 'beat'],
  ['dWnJw60mig0', 'SMLT beats Society', 'beat'], ['Jiby_l51wpA', 'SMLT beats The Golden', 'beat'],
  ['ThvJhbEtvyQ', 'SMLT beats Firework', 'beat'], ['EBGoVlwTieA', 'SMLT beats Tidalbaeb', 'beat'],
  ['AL-c39hXPcU', '[SMLT] Clutter', 'project'], ['KSHADY9mD4o', '[SMLT] PACMAN', 'project'],
  ['hxFDFHR-U_8', '[SMLT] Hopes and Dream', 'project'], ['jkxZe_CPQFQ', '[SMLT] Caterpillar Blitz', 'project'],
  ['dTHVXborVWs', '[SMLT] Rumn Bass', 'project'], ['aTxro5xLt64', '[SMLT] METPO', 'project'],
  ['KErAay1xmRY', '[SMLT] Parfait', 'project'], ['1HsFMyBgvZU', '[SMLT] C TObOY', 'project'],
  ['uPcrGc1Gn1Q', '[SMLT] Jack', 'project'], ['D3AIJIanoTo', '[SMLT] Ne Rofl Collab', 'project'],
  ['0fKT9D63UFE', '[SMLT] Nuclear Fusion', 'project'], ['7UcYhRcM6e4', '[SMLT] Pronyx Code', 'project'],
  ['9jXskpaV4PI', 'Nuclear Fusion by SMLT', 'project'], ['REHYPVCCS_A', '[SMLT] Nuclear Fusion update', 'project'],
  ['mdAmi4JTGjo', '[SMLT] Nuclear Fusion', 'project'], ['wmzNH3Ln9WU', '[SMLT] METRO', 'project'],
]
