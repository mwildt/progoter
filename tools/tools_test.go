package tools

import (
	"encoding/json"
	"os"
	"os/exec"
	"testing"
	"time"
)

// TestStopProcess testet das stop_process Tool
func TestStopProcess(t *testing.T) {
	// Starte einen einfachen Prozess (z.B. sleep)
	cmd := exec.Command("sleep", "10")
	err := cmd.Start()
	if err != nil {
		t.Fatalf("Fehler beim Starten des Testprozesses: %v", err)
	}

	pid := cmd.Process.Pid
	
	// Warte kurz, um sicherzustellen, dass der Prozess läuft
	time.Sleep(1 * time.Second)

	// Teste das StopProcess Tool
	args, err := json.Marshal(StopProcessArgs{PID: pid})
	if err != nil {
		t.Fatalf("Fehler beim Marshalen der Argumente: %v", err)
	}

	_, err = StopProcess(string(args))
	if err != nil {
		t.Fatalf("Fehler beim Beenden des Prozesses: %v", err)
	}

	// Überprüfe, ob der Prozess wirklich beendet wurde
	process, err := os.FindProcess(pid)
	if err != nil {
		t.Fatalf("Fehler beim Suchen des Prozesses: %v", err)
	}

	// Versuche, den Prozess zu signalisieren, um zu sehen, ob er noch läuft
	err = process.Signal(os.Interrupt)
	if err != nil {
		// Wenn der Prozess nicht mehr läuft, ist das erwartet
		if err.Error() == "os: process already finished" {
			t.Log("Prozess wurde erfolgreich beendet")
		} else {
			t.Fatalf("Unbekannter Fehler: %v", err)
		}
	}
}