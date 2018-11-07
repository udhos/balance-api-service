package main

func compareSets(set1, set2 []string) (only1, only2, both []string) {
	tab1 := sliceToMap(set1)
	tab2 := sliceToMap(set2)

	for s := range tab1 {
		if _, found := tab2[s]; found {
			both = append(both, s)
		} else {
			only1 = append(only1, s)
		}
	}

	for s := range tab2 {
		if _, found := tab1[s]; !found {
			only2 = append(only2, s)
		}
	}

	return
}

func sliceToMap(slice []string) map[string]struct{} {
	tab := map[string]struct{}{}
	for _, s := range slice {
		tab[s] = struct{}{}
	}
	return tab
}
