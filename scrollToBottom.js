// Funktion zum automatischen Scrollen nach unten, wenn neue Nachrichten hinzugefügt werden
function scrollToBottomIfAtBottom() {
    // Überprüfen, ob der Benutzer sich am unteren Ende des Bildschirms befindet
    const isAtBottom = (window.innerHeight + window.scrollY) >= document.body.offsetHeight - 10;
    
    // Wenn der Benutzer am unteren Ende ist, nach unten scrollen
    if (isAtBottom) {
        window.scrollTo({ 
            top: document.body.scrollHeight, 
            behavior: 'smooth' 
        });
    }
}

// Beispiel: Diese Funktion kann aufgerufen werden, wenn neue Nachrichten hinzugefügt werden
// z. B. in einer Chat-Anwendung:
// document.getElementById('messages').addEventListener('DOMNodeInserted', scrollToBottomIfAtBottom);

// Exportieren der Funktion, falls sie in einem Modul verwendet wird
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { scrollToBottomIfAtBottom };
}