const input=document.querySelector('#profile-search');
const country=document.querySelector('#country-filter');
const sort=document.querySelector('#profile-sort');
const container=document.querySelector('#player-cards');
const cards=[...document.querySelectorAll('.player-card')];
function update(){
  const q=input.value.trim().toLowerCase(), selected=country.value;
  cards.forEach(card=>card.hidden=!!((q&&!card.dataset.search.includes(q))||(selected&&card.dataset.country!==selected)));
  const ordered=[...cards].sort((a,b)=>sort.value==='points'?Number(b.dataset.points)-Number(a.dataset.points):sort.value==='name'?a.dataset.name.localeCompare(b.dataset.name):Number(a.dataset.rank)-Number(b.dataset.rank));
  ordered.forEach(card=>container.append(card));
}
[input,country,sort].forEach(control=>control?.addEventListener(control===input?'input':'change',update));
