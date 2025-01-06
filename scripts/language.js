document.addEventListener('DOMContentLoaded', () => {
    setupLanguageButtons();
    applyLanguageChoice();
});

function setupLanguageButtons() {
    document.querySelectorAll('.language-selector button').forEach(button => {
        button.addEventListener('click', function() {
            var language = button.textContent.toLowerCase();
            saveLanguageChoice(language);
            applyLanguageChoice();
        });
    });
}

function saveLanguageChoice(language) {
    localStorage.setItem('preferredLanguage', language);
    window.location.reload(); // Optioneel: Herlaad de pagina om de wijziging door te voeren
}

// Functie om de taalvoorkeur toe te passen op de website.
function applyLanguageChoice() {
    var storedLanguage = localStorage.getItem('preferredLanguage') || 'en'; // Standaard naar 'en' als er niks is opgeslagen
    loadLanguageFile(storedLanguage);
}

// Functie om het JSON-bestand met vertalingen te laden en toe te passen.
function loadLanguageFile(language) {
    var filePath = '/locales/' + language + '.json'; // Update dit pad indien nodig.
    fetch(filePath)
        .then(response => response.json())
        .then(data => translatePage(data))
        .catch(error => console.error('Error loading language file:', error));
}

// Functie om de pagina te vertalen met de geladen taalgegevens.
function translatePage(data) {
    document.querySelectorAll('[data-key]').forEach(element => {
        var key = element.getAttribute('data-key');
        if (data[key]) {
            element.innerText = data[key]; // Pas de tekst van het element aan.
        }
    });
    document.querySelector('html').setAttribute('lang', localStorage.getItem('preferredLanguage')); // Update de lang attribuut van de html-tag.
}