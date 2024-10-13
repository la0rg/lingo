function closeOnClick(element) {
    document.addEventListener('click', function (event) {
        element.removeAttribute('open');
    });
}

function addEventListenerToLanguageSelector() {
    closeOnClick(document.getElementById("source-language"));
    closeOnClick(document.getElementById("target-language"));
}

window.onload = function () {
    addEventListenerToLanguageSelector();
};
