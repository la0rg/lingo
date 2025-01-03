function copyAPKG() {
    const apkgBtn = document.getElementById("apkg-download");
    if (!apkgBtn) {
        console.error(`APKG URL not found.`);
        return;
    }
    
    navigator.clipboard.writeText(apkgBtn.href)
    .catch(err => { console.error("Failed to copy APKG URL to clipboard:", err); });
}