document.addEventListener('DOMContentLoaded', () => {
    const localStorageKey = 'readMinifluxEntryIDs';

    const getReadEntryIDs = () => {
        const ids = localStorage.getItem(localStorageKey);
        return ids ? JSON.parse(ids) : [];
    };

    const addReadEntryID = (id) => {
        const ids = getReadEntryIDs();
        if (!ids.includes(id)) {
            ids.push(id);
            localStorage.setItem(localStorageKey, JSON.stringify(ids));
        }
    };

    const removeReadEntryID = (id) => {
        let ids = getReadEntryIDs();
        ids = ids.filter(i => i !== id);
        localStorage.setItem(localStorageKey, JSON.stringify(ids));
    };

    const entries = Array.from(document.querySelectorAll('section.entry'));
    let activeEntry = null;
    let markAsReadTimer = null;
    let focusTimer = null;
    let lastScrollY = window.scrollY;

    function getVisibleEntries() {
        return entries.filter(entry => !entry.classList.contains('entry-hidden'));
    }

    function markAsRead(entry) {
        const visibleEntries = getVisibleEntries();
        const entryIndex = visibleEntries.indexOf(entry);
        for (let i = 0; i < entryIndex; i++) {
            visibleEntries[i].classList.add('viewed');
            addReadEntryID(visibleEntries[i].dataset.entryId);
        }
    }

    function setActiveEntry(entry, preventScroll = false) {
        clearTimeout(markAsReadTimer);

        if (activeEntry) {
            activeEntry.classList.remove('active');
        }

        activeEntry = entry;

        if (activeEntry) {
            activeEntry.classList.add('viewed'); // Mark as viewed immediately
            addReadEntryID(activeEntry.dataset.entryId);
            activeEntry.classList.add('active');
            if (!preventScroll) { // Only focus if not preventing scroll
                activeEntry.querySelector('summary')?.focus();
            }

            // Mark previous entries as read after a delay
            markAsReadTimer = setTimeout(() => {
                markAsRead(activeEntry);
            }, 1000);
        }
    }

    function focusEntry(startIndex, direction) {
        let nextIndex = startIndex;

        const visibleEntries = getVisibleEntries();
        if (visibleEntries.length === 0) {
            setActiveEntry(null);
            return;
        }

        if (direction === 'next') {
            nextIndex = Math.min(startIndex + 1, visibleEntries.length - 1);
        } else if (direction === 'prev') {
            nextIndex = Math.max(startIndex - 1, 0);
        }
        
        const newEntry = visibleEntries[nextIndex];

        if (newEntry) {
            setActiveEntry(newEntry);
        }
    }

    function navigate(direction) {
        const visibleEntries = getVisibleEntries();
        if (visibleEntries.length === 0) return;
        
        const currentIndex = activeEntry ? visibleEntries.indexOf(activeEntry) : -1;

        switch (direction) {
            case 'next':
                focusEntry(currentIndex, 'next');
                break;
            case 'prev':
                focusEntry(currentIndex, 'prev');
                break;
            case 'first':
                setActiveEntry(visibleEntries[0]);
                break;
            case 'last':
                setActiveEntry(visibleEntries[visibleEntries.length - 1]);
                break;
        }
    }

    const observer = new IntersectionObserver((intersectionEntries) => {
        const scrollDirection = window.scrollY > lastScrollY ? 'down' : 'up';
        lastScrollY = window.scrollY;

        // Mark entries as viewed when they scroll out of view
        intersectionEntries.forEach(intersectionEntry => {
            const {
                target,
                isIntersecting,
                boundingClientRect
            } = intersectionEntry;
            if (scrollDirection === 'down' && !isIntersecting && boundingClientRect.top < 0) {
                target.classList.add('viewed');
            }
        });

        // If we are scrolling up, do nothing with focus.
        if (scrollDirection === 'up') {
            return;
        }

        // Check if the active entry is still visible. If so, do nothing.
        if (activeEntry) {
            const rect = activeEntry.getBoundingClientRect();
            if (rect.bottom > 0 && rect.top < window.innerHeight) {
                return;
            }
        }

        // Find the next entry to focus.
        // It should be the first one that is at least partially visible.
        for (const entry of getVisibleEntries()) {
            const rect = entry.getBoundingClientRect();
            if (rect.top >= 0 && rect.top < window.innerHeight) {
                // Only set active if it's after the current active entry.
                const activeEntryIndex = activeEntry ? entries.indexOf(activeEntry) : -1;
                const newEntryIndex = entries.indexOf(entry);
                if (newEntryIndex > activeEntryIndex) {
                    setActiveEntry(entry, true);
                    return;
                }
            }
        }

    }, {
        threshold: 0.0,
    });

    entries.forEach(entry => observer.observe(entry));

    // Restore read state from localStorage on page load
    const readEntryIDsOnLoad = getReadEntryIDs();
    entries.forEach(entry => {
        const entryId = entry.dataset.entryId;
        if (entryId && readEntryIDsOnLoad.includes(entryId)) {
            entry.classList.add('viewed');
        }
    });

    if (entries.length) {
        setActiveEntry(getVisibleEntries()[0]);
    }

    document.addEventListener('click', (event) => {
        const target = event.target;
        const entry = target.closest('section.entry');

        if (entry) {
            setActiveEntry(entry);
        }
    });

    document.addEventListener('keydown', (event) => {
        if (['j', 'k', 'v', 'D', 'U', 'V'].includes(event.key)) {
            event.preventDefault();

            switch (event.key) {
                case 'j':
                    navigate('next');
                    break;
                case 'k':
                    navigate('prev');
                    break;
                case 'U':
                    navigate('last');
                    break;
                case 'D':
                    navigate('first');
                    break;
                case 'V':
                    if (activeEntry) {
                        const link = activeEntry.querySelector('a[data-link-internal]');
                        if (link && link.href) window.open(link.href, '_blank');
                    }
                    break;
                case 'v':
                    if (activeEntry) {
                        const link = activeEntry.querySelector('a[data-link-external]');
                        if (link && link.href) window.open(link.href, '_blank');
                    }
                    break;
            }
        }
    });
});