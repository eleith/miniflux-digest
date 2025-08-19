document.addEventListener('DOMContentLoaded', () => {
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

                intersectionEntries.forEach(intersectionEntry => {
                    const {
                        target,
                        isIntersecting,
                        boundingClientRect
                    } = intersectionEntry;

                    // If an entry is no longer intersecting and we are scrolling down,
                    // mark it as viewed.
                    if (scrollDirection === 'down' && !isIntersecting && boundingClientRect.top < 0) {
                        target.classList.add('viewed');
                    }

                    // If there's no active entry, or the active entry is no longer visible,
                    // try to set the first visible entry as active.
                    if (!activeEntry || (activeEntry && !intersectionEntries.some(e => e.target === activeEntry && e.isIntersecting))) {
                        for (const entry of getVisibleEntries()) {
                            const rect = entry.getBoundingClientRect();
                            // Set the first entry that is at or above the top of the viewport as active.
                            // Only set active if it's not already the active entry and it's not behind the current active entry.
                            if (rect.top >= 0) {
                                if (!activeEntry || entries.indexOf(entry) >= entries.indexOf(activeEntry)) {
                                    setActiveEntry(entry, true); // Prevent scroll on auto-focus
                                }
                                return;
                            }
                        }
                    }
                });
            }, {
                threshold: 0.01 // A very small threshold to detect entry entering/leaving the viewport
            });

    entries.forEach(entry => observer.observe(entry));

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