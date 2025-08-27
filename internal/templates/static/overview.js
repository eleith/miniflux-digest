document.addEventListener('DOMContentLoaded', () => {
    const markAsReadLinks = document.querySelectorAll('.mark-as-read-link');

    markAsReadLinks.forEach(link => {
        link.addEventListener('click', (event) => {
            event.preventDefault();
            const entry = event.target.closest('section.entry');
            if (entry) {
                entry.classList.toggle('viewed');
                if (entry.classList.contains('viewed')) {
                    event.target.textContent = 'Mark as unread';
                } else {
                    event.target.textContent = 'Mark as read';
                }
            }
        });
    });

    const entries = document.querySelectorAll('.grouped-digests .entry');
    entries.forEach(entry => {
        entry.addEventListener('click', (event) => {
            // Don't do anything if a link in the meta section was clicked
            if (event.target.closest('.entry-meta')) {
                return;
            }

            // Find the main link and navigate to it
            const mainLink = entry.querySelector('a.entry-link');
            if (mainLink) {
                mainLink.click();
            }
        });
        // Add a cursor pointer to the whole card except the meta section
        entry.style.cursor = 'pointer';
    });

    // Remove cursor pointer from the meta section
    const metaSections = document.querySelectorAll('.entry-meta');
    metaSections.forEach(meta => {
        meta.style.cursor = 'auto';
    });
});
