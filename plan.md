# Plan für die Implementierung der Löschfunktion für Chat-Contexts

## 1. Analyse der Anforderungen
- **Ziel:** Ein Button im `chat-context-header` soll das Löschen eines Chat-Contexts ermöglichen.
- **Ablauf:**
  1. Nutzer klickt auf den Lösch-Button im Header.
  2. Bestätigungsdialog wird angezeigt.
  3. Bei Bestätigung wird die Löschanfrage an die API gesendet.
  4. Ein `context-deleted`-Event wird gebubbelt.
  5. Die `ContextList`-Komponente reagiert auf das Event und entfernt den Context aus der Liste.
  6. Ein anderer Context wird automatisch ausgewählt und angezeigt.

## 2. Codebase-Analyse
- **Aktuelle Struktur:**
  - `web/resources/app.js`: Enthält die `ChatApp`-Komponente mit dem Header und den Aktionen für Cancel, Compact und Clear.
  - `web/resources/molecules.js`: Enthält die `ContextList`-Komponente, die die Liste der Contexts verwaltet.
  - **Fehlende Funktionen:**
    - Lösch-Button im Header.
    - Bestätigungsdialog.
    - API-Aufruf zum Löschen eines Contexts.
    - Event-Handling für das `context-deleted`-Event.

## 3. Plan

### Schritt 1: Lösch-Button im Header hinzufügen
- **Datei:** `web/resources/app.js`
- **Änderung:**
  - Füge einen neuen Button mit dem Label "Delete" im `header-actions`-Bereich hinzu.
  - Der Button soll nur aktiv sein, wenn kein Processing stattfindet.
  - Der Button löst eine neue Methode `deleteContext` aus.

### Schritt 2: Bestätigungsdialog implementieren
- **Datei:** `web/resources/app.js`
- **Änderung:**
  - Implementiere eine Methode `confirmDelete`, die einen Bestätigungsdialog anzeigt.
  - Bei Bestätigung wird die `deleteContext`-Methode aufgerufen.

### Schritt 3: API-Aufruf zum Löschen eines Contexts
- **Datei:** `web/resources/app.js`
- **Änderung:**
  - Implementiere die Methode `deleteContext`, die eine DELETE-Anfrage an die API sendet (z. B. `http://localhost:8080/chat/${this.contextId}`).
  - Bei Erfolg wird ein `context-deleted`-Event gebubbelt.

### Schritt 4: Event-Handling in der ContextList
- **Datei:** `web/resources/molecules.js`
- **Änderung:**
  - Füge einen Event-Listener für das `context-deleted`-Event hinzu.
  - Bei Empfang des Events wird der Context aus der Liste entfernt.
  - Ein anderer Context wird automatisch ausgewählt (z. B. der erste in der Liste).

### Schritt 5: Automatische Auswahl eines anderen Contexts
- **Datei:** `web/resources/molecules.js`
- **Änderung:**
  - Nach dem Löschen eines Contexts wird überprüft, ob noch Contexts vorhanden sind.
  - Falls ja, wird der erste Context in der Liste ausgewählt.
  - Falls nein, wird ein neuer Context erstellt oder eine Meldung angezeigt.

## 4. Umsetzung

### Schritt 1: Lösch-Button im Header hinzufügen
```javascript
<atomic-button label="Delete" ?disabled=${this.processing} @button-click=${this.confirmDelete}></atomic-button>
```

### Schritt 2: Bestätigungsdialog implementieren
```javascript
confirmDelete() {
    if (confirm('Are you sure you want to delete this context?')) {
        this.deleteContext();
    }
}
```

### Schritt 3: API-Aufruf zum Löschen eines Contexts
```javascript
async deleteContext() {
    try {
        const response = await fetch(`http://localhost:8080/chat/${this.contextId}`, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
            },
        });
        if (!response.ok) {
            throw new Error('Failed to delete context');
        }
        // Event bubbeln
        this.dispatchEvent(new CustomEvent('context-deleted', {
            detail: { contextId: this.contextId },
            bubbles: true,
            composed: true
        }));
    } catch (error) {
        console.error('Error deleting context:', error);
    }
}
```

### Schritt 4: Event-Handling in der ContextList
```javascript
connectedCallback() {
    super.connectedCallback();
    this.fetchContexts();
    this.addEventListener('context-deleted', this.handleContextDeleted);
}

disconnectedCallback() {
    super.disconnectedCallback();
    this.removeEventListener('context-deleted', this.handleContextDeleted);
}

handleContextDeleted(e) {
    const deletedContextId = e.detail.contextId;
    this.contexts = this.contexts.filter(contextId => contextId !== deletedContextId);
    if (this.contexts.length > 0) {
        this.selectedContext = this.contexts[0];
    } else {
        this.selectedContext = null;
    }
}
```

### Schritt 5: Automatische Auswahl eines anderen Contexts
- Diese Logik ist bereits in Schritt 4 enthalten.

## 5. Verifizierung
- **Tests:**
  - Überprüfe, ob der Lösch-Button angezeigt wird.
  - Teste den Bestätigungsdialog.
  - Teste den API-Aufruf und das Event-Handling.
  - Überprüfe, ob der Context aus der Liste entfernt wird.
  - Überprüfe, ob ein anderer Context automatisch ausgewählt wird.

## 6. Dokumentation
- Aktualisiere die README.md, um die neue Funktion zu beschreiben.
- Füge Kommentare im Code hinzu, um die Logik zu erklären.
