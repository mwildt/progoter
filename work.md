# Aufgaben für den Umbau der Chat-Kontextverwaltung

## Ziel
Erweiterung des Systems, um neue Chat-Kontexte mit einem angegebenen Arbeitsverzeichnis (workspace-dir) zu erstellen. Die UI soll eine Liste der vorhandenen Chat-Kontexte links anzeigen und rechts den Chat zum ausgewählten Kontext.

## Analyse
- **Aktueller Zustand**:
  - Es gibt einen `ContextManager`, der Chat-Kontexte verwaltet.
  - Die UI zeigt derzeit nur einen Chat-Kontext an (Standard: "default").
  - Chat-Kontexte haben ein `BasePath`, das das Arbeitsverzeichnis darstellt.
  - Die REST-API unterstützt bereits die Verwaltung von Kontexten über IDs.

- **Anforderungen**:
  1. **Neue Kontexte mit workspace-dir**: Es soll möglich sein, neue Chat-Kontexte mit einem benutzerdefinierten Arbeitsverzeichnis zu erstellen.
  2. **UI-Anpassung**:
     - Links: Liste der vorhandenen Chat-Kontexte.
     - Rechts: Chat-Ansicht für den ausgewählten Kontext.
  3. **Backend-Anpassung**:
     - Erweitern des `ContextManager`, um das Arbeitsverzeichnis bei der Erstellung eines Kontexts zu berücksichtigen.
     - Neue REST-Endpunkte für die Erstellung von Kontexten mit Arbeitsverzeichnis.

## Aufgaben

### 1. Backend-Anpassungen

#### 1.1 `ChatContext` erweitern ✅
- **Aufgabe**: Anpassen des `ChatContext`, um das Arbeitsverzeichnis bei der Erstellung zu berücksichtigen.
- **Datei**: `service/chat_context.go`
- **Änderungen**:
  - `NewChatContext` soll ein `basePath`-Parameter akzeptieren.
  - `BasePath` soll bei der Initialisierung gesetzt werden.

#### 1.2 `ContextManager` erweitern ✅
- **Aufgabe**: Erweitern des `ContextManager`, um neue Kontexte mit einem Arbeitsverzeichnis zu erstellen.
- **Datei**: `service/context_manager.go`
- **Änderungen**:
  - `CreateContext` soll einen `basePath`-Parameter akzeptieren und an `NewChatContext` weitergeben.

#### 1.3 REST-Endpunkt für die Erstellung von Kontexten ✅
- **Aufgabe**: Neuen Endpunkt `/chat` (POST) hinzufügen, um einen neuen Kontext mit einem Arbeitsverzeichnis zu erstellen.
- **Datei**: `service/rest_controller.go`
- **Änderungen**:
  - `PostCreateContextHandler`: Handler für die Erstellung eines neuen Kontexts.
  - Der Handler soll eine JSON-Anfrage mit `id` und `basePath` akzeptieren.
  - Der Handler soll den neuen Kontext im `ContextManager` erstellen.

#### 1.4 REST-Endpunkt für die Liste der Kontexte ✅
- **Aufgabe**: Neuen Endpunkt `/chat` (GET) hinzufügen, um eine Liste aller verfügbaren Kontexte zurückzugeben.
- **Datei**: `service/rest_controller.go`
- **Änderungen**:
  - `GetContextsHandler`: Handler für die Rückgabe der Liste der Kontexte.
  - Der Handler soll die IDs der Kontexte als JSON-Array zurückgeben.

### 2. UI-Anpassungen

#### 2.1 UI-Komponente für die Kontextliste
- **Aufgabe**: Neue UI-Komponente `<context-list>` erstellen, die die Liste der Kontexte anzeigt.
- **Datei**: `web/resources/molecules.js`
- **Änderungen**:
  - Komponente soll die Kontext-IDs vom Backend abrufen und als Liste anzeigen.
  - Auswahl eines Kontexts soll ein Event auslösen, das den ausgewählten Kontext an die Chat-Ansicht übergibt.

#### 2.2 UI-Komponente für die Chat-Ansicht
- **Aufgabe**: Anpassen der `<chat-app>`-Komponente, um den ausgewählten Kontext anzuzeigen.
- **Datei**: `web/resources/app.js`
- **Änderungen**:
  - Komponente soll auf das Event der Kontextliste reagieren und den Chat für den ausgewählten Kontext laden.
  - Der SSE-Stream soll dynamisch basierend auf der ausgewählten Kontext-ID aktualisiert werden.

#### 2.3 Layout-Anpassung
- **Aufgabe**: Anpassen des Layouts, um die Kontextliste links und die Chat-Ansicht rechts anzuzeigen.
- **Datei**: `web/resources/index.html` und `web/resources/molecules.js`
- **Änderungen**:
  - `<app-layout>` soll ein Grid-Layout verwenden, um die beiden Komponenten nebeneinander anzuzeigen.
  - `<context-list>` soll links und `<chat-app>` rechts platziert werden.

### 3. Integration und Tests

#### 3.1 Backend-Tests
- **Aufgabe**: Unit-Tests für die neuen Endpunkte und die erweiterten Funktionen schreiben.
- **Datei**: `service/rest_controller_test.go`
- **Änderungen**:
  - Tests für `PostCreateContextHandler` und `GetContextsHandler`.
  - Tests für die Erstellung von Kontexten mit Arbeitsverzeichnis.

#### 3.2 UI-Tests
- **Aufgabe**: Manuelle Tests der UI, um sicherzustellen, dass die Kontextliste und die Chat-Ansicht korrekt funktionieren.
- **Datei**: `web/resources/*`
- **Änderungen**:
  - Überprüfen, ob die Kontextliste korrekt geladen wird.
  - Überprüfen, ob der Chat für den ausgewählten Kontext korrekt angezeigt wird.

## Meilensteine

1. **Backend-Anpassungen**:
   - `ChatContext` und `ContextManager` erweitern.
   - Neue REST-Endpunkte implementieren.
   - Unit-Tests schreiben.

2. **UI-Anpassungen**:
   - `<context-list>`-Komponente erstellen.
   - `<chat-app>`-Komponente anpassen.
   - Layout anpassen.

3. **Integration und Tests**:
   - Backend- und UI-Tests durchführen.
   - Manuelle Tests der gesamten Anwendung.

## Zeitplan
- **Tag 1**: Backend-Anpassungen und Unit-Tests.
- **Tag 2**: UI-Anpassungen und manuelle Tests.
- **Tag 3**: Integration und finale Tests.
