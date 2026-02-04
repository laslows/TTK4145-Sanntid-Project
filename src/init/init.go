package initialize

import (
	"fmt"
)

func Initialize() {
	//opprette kontakt, finne ut hva slags rolle du har 
	//(hvis det allerede er en master i nettverket, blir du slave. 
	// Hvis du er den eneste heisen i nettverket blir du master, 
	// hvis to mastere merges sammen, 
	// eller hvis det ikke finnes en master i nettverket, 
	// brukes en enkel regel 
	// (f.eks. lavest IP-adresse eller heis-ID) 
	// for å bestemme hvem av de som skal være master, 
	// og hvem som skal være slave.
	fmt.Print("Initialiser heisen")
}