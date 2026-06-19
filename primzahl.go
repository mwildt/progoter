package main

// IstPrimzahl prüft, ob eine gegebene Zahl eine Primzahl ist.
func IstPrimzahl(n int) bool {
	if n <= 1 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}
