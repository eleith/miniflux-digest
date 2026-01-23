document.addEventListener('DOMContentLoaded', () => {
    const digestDate = document.body.dataset.digestDate;
    const localStorageKey = `readSubgroupSlugs_${digestDate}`;

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

    const applyReadStatus = () => {
        const readSlugs = getReadSlugs();
        document.querySelectorAll('section.entry').forEach(entry => {
            const slug = entry.dataset.slug;
            const markAsReadLink = entry.querySelector('.mark-as-read-link');

            if (slug && readSlugs.includes(slug)) {
                entry.classList.add('viewed');
                if (markAsReadLink) {
                    markAsReadLink.textContent = 'Mark as unread';
                }
            } else {
                entry.classList.remove('viewed');
                if (markAsReadLink) {
                    markAsReadLink.textContent = 'Mark as read';
                }
            }
        });
    };

    applyReadStatus();

    window.addEventListener('pageshow', () => {
        applyReadStatus();
    });

    const markAsReadLinks = document.querySelectorAll('.mark-as-read-link');
    markAsReadLinks.forEach(link => {
        link.addEventListener('click', (event) => {
            event.preventDefault();
            event.stopPropagation();
            
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

    const entries = document.querySelectorAll('.grouped-digests .entry');
    entries.forEach(entry => {
        entry.addEventListener('click', (event) => {
            if (event.target.closest('.mark-as-read-link')) {
                return;
            }

            const slug = entry.dataset.slug;
            const mainLink = entry.querySelector('a.entry-link');

            if (slug) {
                addReadSlug(slug);
                entry.classList.add('viewed');
                const markBtn = entry.querySelector('.mark-as-read-link');
                if (markBtn) markBtn.textContent = 'Mark as unread';
            }

            if (mainLink) {
                if (event.target.closest('a.entry-link')) {
                    return;
                }

                window.location.href = mainLink.href;
            }
        });

        entry.style.cursor = 'pointer';
    });

    const metaSections = document.querySelectorAll('.entry-meta');
    metaSections.forEach(meta => {
        meta.style.cursor = 'auto';
    });
});
