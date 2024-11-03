function closeOnClick(element) {
    document.addEventListener('click', function (event) {
        if (!element.contains(event.target)) {
            element.removeAttribute('open');
        }
    });
}

function addEventListenerToLanguageSelector(content) {
    var langSelectors = content.querySelectorAll("#source-language, #target-language");
    for (var i = 0; i < langSelectors.length; i++) {
        closeOnClick(langSelectors[i]);
    }
}

htmx.onLoad(function (content) {
    addEventListenerToLanguageSelector(content);
});