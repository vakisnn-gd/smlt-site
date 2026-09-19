package main

import "log"

var defaultEvents = []Event{
	{VideoID: "axsCwtl2Hro", Title: "SMLT beats Back on Track", Category: "beat"},
	{VideoID: "58ItHCC8ZKI", Title: "SMLT beats Every End", Category: "beat"},
	{VideoID: "8RduRoZmFHs", Title: "SMLT beats Ne Rofl Collab", Category: "beat"},
	{VideoID: "dWnJw60mig0", Title: "SMLT beats Society", Category: "beat"},
	{VideoID: "Jiby_l51wpA", Title: "SMLT beats The Golden", Category: "beat"},
	{VideoID: "ThvJhbEtvyQ", Title: "SMLT beats Firework", Category: "beat"},
	{VideoID: "EBGoVlwTieA", Title: "SMLT beats Tidalbaeb", Category: "beat"},
	{VideoID: "AL-c39hXPcU", Title: "[SMLT] Clutter", Category: "project"},
	{VideoID: "KSHADY9mD4o", Title: "[SMLT] PACMAN", Category: "project"},
	{VideoID: "hxFDFHR-U_8", Title: "[SMLT] Hopes and Dream", Category: "project"},
	{VideoID: "jkxZe_CPQFQ", Title: "[SMLT] Caterpillar Blitz", Category: "project"},
	{VideoID: "dTHVXborVWs", Title: "[SMLT] Rumn Bass", Category: "project"},
	{VideoID: "aTxro5xLt64", Title: "[SMLT] METPO", Category: "project"},
	{VideoID: "KErAay1xmRY", Title: "[SMLT] Parfait", Category: "project"},
	{VideoID: "1HsFMyBgvZU", Title: "[SMLT] C TObOY", Category: "project"},
	{VideoID: "uPcrGc1Gn1Q", Title: "[SMLT] Jack", Category: "project"},
	{VideoID: "D3AIJIanoTo", Title: "[SMLT] Ne Rofl Collab", Category: "project"},
	{VideoID: "0fKT9D63UFE", Title: "[SMLT] Nuclear Fusion", Category: "project"},
	{VideoID: "7UcYhRcM6e4", Title: "[SMLT] Pronyx Code", Category: "project"},
	{VideoID: "9jXskpaV4PI", Title: "Nuclear Fusion by SMLT", Category: "project"},
	{VideoID: "REHYPVCCS_A", Title: "[SMLT] Nuclear Fusion update", Category: "project"},
	{VideoID: "mdAmi4JTGjo", Title: "[SMLT] Nuclear Fusion", Category: "project"},
	{VideoID: "wmzNH3Ln9WU", Title: "[SMLT] METRO", Category: "project"},
}

func seedEventsIfEmpty() {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM events").Scan(&count); err != nil || count > 0 {
		return
	}
	tx, err := db.Begin()
	if err != nil {
		return
	}
	for order, event := range defaultEvents {
		if _, err := tx.Exec("INSERT INTO events (video_id, title, category, sort_order) VALUES ($1,$2,$3,$4)", event.VideoID, event.Title, event.Category, order); err != nil {
			tx.Rollback()
			return
		}
	}
	if err := tx.Commit(); err != nil {
		return
	}
	log.Printf("[DB] Seeded %d events", len(defaultEvents))
}

func seedPlayers() {
	seeds := []Player{
		{Country: "RS", Name: "SpaceRS", Points: 1713.80, Demon: "Nullscapes", GlobalRank: 139},
		{Country: "RU", Name: "Florned", Points: 1643.07, Demon: "Defeated Circles", GlobalRank: 146},
		{Country: "RU", Name: "Flik", Points: 866.17, Demon: "Firework", GlobalRank: 229},
		{Country: "RU", Name: "npoctou_gamer", Points: 287.81, Demon: "Sevvend Clubstep", GlobalRank: 522},
		{Country: "KZ", Name: "euphoria", Points: 235.95, Demon: "poocubed", GlobalRank: 565},
		{Country: "RU", Name: "Tikys", Points: 186.87, Demon: "Sevvend Clubstep", GlobalRank: 599},
		{Country: "OTHER", Name: "CandyCloud22", Points: 120.70, Demon: "Cognition", GlobalRank: 716},
		{Country: "RU", Name: "toxik blaze", Points: 102.05, Demon: "Tartarus", GlobalRank: 766},
		{Country: "RU", Name: "EfzEnn", Points: 92.21, Demon: "Oblivion", GlobalRank: 804},
		{Country: "BG", Name: "tapxyhh", Points: 35.01, Demon: "UNKNOWN", GlobalRank: 1221},
		{Country: "RU", Name: "Leeya", Points: 12.63, Demon: "Arctic Lights", GlobalRank: 1858},
		{Country: "RU", Name: "kocheryzhka", Points: 11.99, Demon: "Bloodlust", GlobalRank: 1899},
		{Country: "BY", Name: "ramp1941", Points: 11.34, Demon: "shimmer", GlobalRank: 1961},
		{Country: "RU", Name: "samoletik", Points: 10.63, Demon: "Sink", GlobalRank: 2022},
		{Country: "RU", Name: "vv4zd", Points: 9.98, Demon: "Renevant", GlobalRank: 2078},
		{Country: "DE", Name: "yeahme", Points: 7.10, Demon: "Sonic Wave", GlobalRank: 2450},
		{Country: "UA", Name: "Vakum", Points: 7.07, Demon: "Cobwebs", GlobalRank: 2457},
		{Country: "RU", Name: "Linqwq", Points: 6.52, Demon: "Wasureta", GlobalRank: 2543},
		{Country: "RU", Name: "H30n41k_GmD", Points: 6.46, Demon: "Sink", GlobalRank: 2561},
		{Country: "AM", Name: "SerGio", Points: 6.33, Demon: "Wasureta", GlobalRank: 2596},
		{Country: "RU", Name: "Спини", Points: 5.85, Demon: "RUST", GlobalRank: 2690},
		{Country: "RU", Name: "RossceorpGD", Points: 5.70, Demon: "ZAPHKIEL", GlobalRank: 2736},
		{Country: "KZ", Name: "69liqu69", Points: 5.10, Demon: "Molten Core", GlobalRank: 2880},
		{Country: "UA", Name: "Imdrinkingtea", Points: 4.82, Demon: "Quantum Processing", GlobalRank: 2969},
		{Country: "UA", Name: "dugen", Points: 4.53, Demon: "Ghoul", GlobalRank: 3071},
		{Country: "RU", Name: "kotacub", Points: 4.40, Demon: "Golden Club", GlobalRank: 3124},
		{Country: "RU", Name: "KotKartofel", Points: 4.32, Demon: "Pulsar", GlobalRank: 3151},
		{Country: "RU", Name: "NopanicGD", Points: 3.52, Demon: "Diamond Disco", GlobalRank: 3512},
		{Country: "RU", Name: "NatrixGMD", Points: 3.06, Demon: "Congregation", GlobalRank: 3808},
		{Country: "RS", Name: "zerrga", Points: 2.24, Demon: "Quantum Processing", GlobalRank: 4467},
		{Country: "RU", Name: "paradoxiz", Points: 1.82, Demon: "INNARDS", GlobalRank: 4940},
		{Country: "RU", Name: "toxatort", Points: 1.72, Demon: "Blade of Justice", GlobalRank: 5055},
		{Country: "UA", Name: "fottex", Points: 1.47, Demon: "Sonic Wave", GlobalRank: 5439},
		{Country: "RU", Name: "Marzyiiik", Points: 1.19, Demon: "Blade of Justice", GlobalRank: 6003},
		{Country: "UA", Name: "Daggit", Points: 1.02, Demon: "Overtime", GlobalRank: 6359},
		{Country: "UA", Name: "KasaneTeto", Points: 1.01, Demon: "Heavens Door", GlobalRank: 6377},
		{Country: "RU", Name: "itzslxnq", Points: 0.71, Demon: "Shinigami", GlobalRank: 7500},
		{Country: "RU", Name: "aerongmd", Points: 0.63, Demon: "Bloodbath", GlobalRank: 8042},
		{Country: "RU", Name: "Заварррка", Points: 0.58, Demon: "Anahita", GlobalRank: 8372},
		{Country: "RU", Name: "matveypro13", Points: 0.36, Demon: "Hopping Over Puddles", GlobalRank: 9608},
		{Country: "RU", Name: "Roflin", Points: 0.31, Demon: "Cataclysm", GlobalRank: 9848},
		{Country: "RU", Name: "Filkoty", Points: 0.13, Demon: "Me Lin A", GlobalRank: 11310},
		{Country: "BY", Name: "Denchis", Points: 0.13, Demon: "Cataclysm", GlobalRank: 11473},
		{Country: "RU", Name: "Fanim59", Points: 0.12, Demon: "Make It Drop", GlobalRank: 11787},
		{Country: "UA", Name: "prostoymofficial", Points: 0.07, Demon: "Acu", GlobalRank: 12308},
		{Country: "RU", Name: "DarBeast", Points: 0.07, Demon: "Acu", GlobalRank: 12328},
		{Country: "OTHER", Name: "CharaGMDq", Points: 0.00, Demon: "—", GlobalRank: 0},
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal("Failed to begin transaction:", err)
	}

	stmt, err := tx.Prepare("INSERT INTO players (rank_order, country, name, points, demon, global_rank) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		log.Fatal("Failed to prepare seed statement:", err)
	}

	for i, p := range seeds {
		stmt.Exec(i+1, p.Country, p.Name, p.Points, p.Demon, p.GlobalRank)
	}
	stmt.Close()
	tx.Commit()
	log.Printf("[DB] Seeded %d players", len(seeds))
}
