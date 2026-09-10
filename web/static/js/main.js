// Custom Alpine.js components
document.addEventListener('alpine:init', () => {
    Alpine.data('counter', () => ({
        count: 0,
        target: 100,
        init() {
            let step = Math.ceil(this.target / 50);
            let interval = setInterval(() => {
                if (this.count < this.target) {
                    this.count += step;
                } else {
                    this.count = this.target;
                    clearInterval(interval);
                }
            }, 30);
        }
    }));
});

function updateActiveNavigation() {
    const currentPath = window.location.pathname.replace(/\/$/, '') || '/';
    document.querySelectorAll('.desktop-nav a, .mobile-nav a').forEach((link) => {
        const linkPath = new URL(link.href, window.location.origin).pathname.replace(/\/$/, '') || '/';
        const isActive = linkPath === currentPath;
        link.classList.toggle('is-active', isActive);
        if (isActive) {
            link.setAttribute('aria-current', 'page');
        } else {
            link.removeAttribute('aria-current');
        }
    });
}

function refreshIcons() {
    if (window.lucide && typeof window.lucide.createIcons === 'function') {
        window.lucide.createIcons();
    }
}

function initProjectNumbers() {
    document.querySelectorAll('[data-project-number]').forEach((number) => {
        const index = Number(number.dataset.projectNumber);
        if (!Number.isNaN(index)) {
            number.textContent = String(index + 1).padStart(2, '0').replace(/\d/g, (digit) => '۰۱۲۳۴۵۶۷۸۹'[digit]);
        }
    });
}

function toPersianDigits(value) {
    return String(value).replace(/\d/g, (digit) => '۰۱۲۳۴۵۶۷۸۹'[digit]);
}

function initAboutCounters() {
    document.querySelectorAll('[data-counter]').forEach((counter) => {
        if (counter.dataset.counterBound === 'true') return;
        counter.dataset.counterBound = 'true';
        const target = Number(counter.dataset.counter);
        const render = (value) => { counter.textContent = toPersianDigits(value); };
        render(0);
        if (!('IntersectionObserver' in window)) {
            render(target);
            return;
        }
        const observer = new IntersectionObserver((entries, currentObserver) => {
            entries.forEach((entry) => {
                if (!entry.isIntersecting) return;
                const startedAt = performance.now();
                const duration = 1500;
                const tick = (now) => {
                    const progress = Math.min((now - startedAt) / duration, 1);
                    const eased = 1 - Math.pow(1 - progress, 3);
                    render(Math.round(target * eased));
                    if (progress < 1) requestAnimationFrame(tick);
                };
                requestAnimationFrame(tick);
                currentObserver.unobserve(counter);
            });
        }, { threshold: 0.65 });
        observer.observe(counter);
    });
}

function initAboutAtomField() {
    const hero = document.querySelector('.about-hero');
    if (!hero || hero.dataset.atomsBound === 'true') return;
    hero.dataset.atomsBound = 'true';
    const atoms = hero.querySelectorAll('.about-atoms span');
    hero.addEventListener('pointermove', (event) => {
        const bounds = hero.getBoundingClientRect();
        const x = (event.clientX - bounds.left) / bounds.width - 0.5;
        const y = (event.clientY - bounds.top) / bounds.height - 0.5;
        atoms.forEach((atom, index) => {
            const strength = (index % 3 + 1) * 14;
            atom.style.transform = `translate(${x * strength}px, ${y * strength}px)`;
        });
    });
    hero.addEventListener('pointerleave', () => {
        atoms.forEach((atom) => { atom.style.transform = ''; });
    });
}

function initScrollReveal() {
    const revealItems = document.querySelectorAll('.value-section, .ventures-section, .packages-section, .portfolio-section, .stats-section, .contact-section, .venture-card, .project-feature, .project-item, .brand-card, .about-page .scroll-reveal, .team-lead, .team-values');
    if (!('IntersectionObserver' in window)) {
        revealItems.forEach((item) => item.classList.add('is-visible'));
        return;
    }
    const observer = new IntersectionObserver((entries, currentObserver) => {
        entries.forEach((entry) => {
            if (entry.isIntersecting) {
                entry.target.classList.add('is-visible');
                currentObserver.unobserve(entry.target);
            }
        });
    }, { threshold: 0.12 });
    revealItems.forEach((item) => {
        item.classList.add('scroll-reveal');
        observer.observe(item);
    });
}

function initVentureCards() {
    document.querySelectorAll('.venture-card__toggle').forEach((toggle) => {
        if (toggle.dataset.bound === 'true') return;
        toggle.dataset.bound = 'true';
        toggle.addEventListener('click', () => {
            const card = toggle.closest('.venture-card');
            const isExpanded = card.classList.toggle('is-expanded');
            toggle.setAttribute('aria-expanded', String(isExpanded));
            toggle.firstChild.textContent = isExpanded ? 'بستن جزئیات ' : 'نمایش جزئیات ';
            if (window.lucide) lucide.createIcons();
        });
    });
}

function initMobileNavDrag() {
    document.querySelectorAll('.mobile-nav__handle').forEach((handle) => {
        if (handle.dataset.bound === 'true') return;
        handle.dataset.bound = 'true';
        let startY = 0;
        handle.addEventListener('pointerdown', (event) => {
            startY = event.clientY;
            handle.setPointerCapture(event.pointerId);
        });
        handle.addEventListener('pointerup', (event) => {
            if (event.clientY - startY > 45) {
                const nav = handle.closest('nav');
                const toggle = nav.querySelector('.menu-toggle');
                if (nav.classList.contains('nav-sheet-open')) toggle.click();
            }
        });
    });
}

function initMobileMenu() {
    document.querySelectorAll('.site-nav').forEach((nav) => {
        if (nav.dataset.menuBound === 'true') return;
        nav.dataset.menuBound = 'true';
        const toggle = nav.querySelector('.menu-toggle');
        const sheet = nav.querySelector('.mobile-nav');
        const close = () => {
            nav.classList.remove('nav-sheet-open');
            toggle.setAttribute('aria-expanded', 'false');
        };
        toggle.addEventListener('click', () => {
            const isOpen = nav.classList.toggle('nav-sheet-open');
            toggle.setAttribute('aria-expanded', String(isOpen));
        });
        sheet.querySelectorAll('a').forEach((link) => link.addEventListener('click', close));
        document.addEventListener('click', (event) => {
            if (!nav.contains(event.target)) close();
        });
    });
}

function initScrollHeader() {
    const header = document.querySelector('.site-header');
    const hero = document.querySelector('.hero-section');
    const spacer = document.querySelector('.header-spacer');
    if (!header || !spacer) return;
    const updateHeader = () => {
        const shouldFix = hero
            ? hero.getBoundingClientRect().bottom <= 0
            : window.scrollY > 80;
        header.classList.toggle('site-header--fixed', shouldFix);
        spacer.classList.toggle('is-active', shouldFix);
    };
    if (header.dataset.scrollBound !== 'true') {
        header.dataset.scrollBound = 'true';
        window.addEventListener('scroll', updateHeader, { passive: true });
    }
    updateHeader();
}

document.addEventListener('DOMContentLoaded', () => {
    refreshIcons();
    initProjectNumbers();
    initAboutCounters();
    initAboutAtomField();
    updateActiveNavigation();
    initScrollReveal();
    initVentureCards();
    initMobileNavDrag();
    initMobileMenu();
    initScrollHeader();
});
document.addEventListener('htmx:afterSettle', () => {
    refreshIcons();
    initProjectNumbers();
    initAboutCounters();
    initAboutAtomField();
    updateActiveNavigation();
    initScrollReveal();
    initVentureCards();
    initMobileNavDrag();
    initMobileMenu();
    initScrollHeader();
});
window.addEventListener('popstate', updateActiveNavigation);
