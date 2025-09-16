document.addEventListener('DOMContentLoaded', () => {
    const digestDate = document.body.dataset.digestDate;
    const localStorageKey = `readSubgroupSlugs_${digestDate}`;

    // Helper functions for localStorage
    const getReadSlugs = () => {
        const slugs = localStorage.getItem(localStorageKey);
        return slugs ? JSON.parse(slugs) : [];
    };

    const addReadSlug = (slug) => {
        const slugs = getReadSlugs();
        if (!slugs.includes(slug)) {
            slugs.push(slug);
            localStorage.setItem(localStorageKey, JSON.stringify(slugs));
        }
    };

    const removeReadSlug = (slug) => {
        let slugs = getReadSlugs();
        slugs = slugs.filter(s => s !== slug);
        localStorage.setItem(localStorageKey, JSON.stringify(slugs));
    };


    const markAsReadLinks = document.querySelectorAll('.mark-as-read-link');

    markAsReadLinks.forEach(link => {
        link.addEventListener('click', (event) => {
            event.preventDefault();
            const entry = event.target.closest('section.entry');
            if (entry) {
                const slug = entry.dataset.slug;

                entry.classList.toggle('viewed');
                if (entry.classList.contains('viewed')) {
                    event.target.textContent = 'Mark as unread';
                    if (slug) {
                        addReadSlug(slug);
                    }
                } else {
                    event.target.textContent = 'Mark as read';
                    if (slug) {
                        removeReadSlug(slug);
                    }
                }
            }
        });
    });

    const applyReadStatusOnLoad = () => {
        const readSlugs = getReadSlugs();
        document.querySelectorAll('section.entry').forEach(entry => {
            const slug = entry.dataset.slug;
            const markAsReadLink = entry.querySelector('.mark-as-read-link');

            if (slug && readSlugs.includes(slug)) {
                entry.classList.add('viewed');
                if (markAsReadLink) {
                    markAsReadLink.textContent = 'Mark as unread';
                }
            }
        });
    };

    applyReadStatusOnLoad(); // Call on page load

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
                const slug = entry.dataset.slug; // Get slug from the parent entry element
                addReadSlug(slug); // Mark as read when navigating
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
