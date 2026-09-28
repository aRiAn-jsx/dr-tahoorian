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
            if (handle.setPointerCapture) {
                try { handle.setPointerCapture(event.pointerId); } catch (_) {}
            }
        });
        handle.addEventListener('pointerup', (event) => {
            if (event.clientY - startY > 45) {
                const toggle = document.querySelector('.menu-toggle');
                if (toggle && (document.body.classList.contains('nav-sheet-open') || document.querySelector('.nav-sheet-open'))) {
                    toggle.click();
                }
            }
        });
    });
}

function initMobileMenu() {
    const nav = document.querySelector('.site-nav');
    const toggle = document.querySelector('.menu-toggle');
    const sheet = document.getElementById('mobile-navigation') || document.querySelector('.mobile-nav');
    const overlay = document.querySelector('.nav-overlay');

    if (!toggle || toggle.dataset.menuBound === 'true') return;
    toggle.dataset.menuBound = 'true';

    const close = () => {
        document.body.classList.remove('nav-sheet-open');
        if (nav) nav.classList.remove('nav-sheet-open');
        toggle.setAttribute('aria-expanded', 'false');
    };

    const open = () => {
        document.body.classList.add('nav-sheet-open');
        if (nav) nav.classList.add('nav-sheet-open');
        toggle.setAttribute('aria-expanded', 'true');
    };

    toggle.addEventListener('click', (e) => {
        e.stopPropagation();
        const isOpen = document.body.classList.contains('nav-sheet-open') || (nav && nav.classList.contains('nav-sheet-open'));
        if (isOpen) {
            close();
        } else {
            open();
        }
    });

    if (overlay && overlay.dataset.menuBound !== 'true') {
        overlay.dataset.menuBound = 'true';
        overlay.addEventListener('click', close);
    }

    if (sheet && sheet.dataset.menuBound !== 'true') {
        sheet.dataset.menuBound = 'true';
        sheet.querySelectorAll('a').forEach((link) => {
            link.addEventListener('click', close);
        });
    }

    if (!window.mobileMenuGlobalBound) {
        window.mobileMenuGlobalBound = true;
        document.addEventListener('keydown', (event) => {
            if (event.key === 'Escape' && document.body.classList.contains('nav-sheet-open')) {
                close();
            }
        });
    }
}

function initScrollHeader() {
    const header = document.querySelector('.site-header');
    const hero = document.querySelector('.hero-section');
    const contactHero = document.querySelector('.contact-hero-section');
    const spacer = document.querySelector('.header-spacer');
    if (!header || !spacer) return;
    const updateHeader = () => {
        const heroSection = hero || contactHero;
        const shouldFix = heroSection
            ? heroSection.getBoundingClientRect().bottom <= 0
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

function initContactPageAnimations() {
    // Hero Particles Animation
    const heroParticles = document.getElementById('hero-particles');
    if (heroParticles) {
        const createParticle = () => {
            const particle = document.createElement('div');
            particle.style.cssText = `
                position: absolute;
                width: ${Math.random() * 4 + 2}px;
                height: ${Math.random() * 4 + 2}px;
                background: rgba(216, 178, 115, ${Math.random() * 0.5 + 0.2});
                border-radius: 50%;
                left: ${Math.random() * 100}%;
                top: ${Math.random() * 100}%;
                pointer-events: none;
                animation: particleFloat ${Math.random() * 10 + 10}s infinite ease-in-out;
                animation-delay: ${Math.random() * 5}s;
            `;
            heroParticles.appendChild(particle);
        };

        for (let i = 0; i < 20; i++) {
            createParticle();
        }
    }

    // Contact Cards Hover Effect
    const contactCards = document.querySelectorAll('.contact-card-item--modern');
    contactCards.forEach(card => {
        card.addEventListener('mouseenter', function() {
            this.style.transform = 'translateY(-5px) scale(1.02)';
        });
        card.addEventListener('mouseleave', function() {
            this.style.transform = '';
        });
    });

    // Form Input Animation
    const formInputs = document.querySelectorAll('.form-field--floating input, .form-field--floating select, .form-field--floating textarea');
    formInputs.forEach(input => {
        input.addEventListener('focus', function() {
            this.parentElement.classList.add('is-focused');
        });
        input.addEventListener('blur', function() {
            this.parentElement.classList.remove('is-focused');
        });
    });

    // Social Chips Animation
    const socialChips = document.querySelectorAll('.social-chip--modern');
    socialChips.forEach(chip => {
        chip.addEventListener('mouseenter', function() {
            this.style.transform = 'translateY(-3px) scale(1.05)';
        });
        chip.addEventListener('mouseleave', function() {
            this.style.transform = '';
        });
    });

    // FAQ Accordion Smooth Animation
    const faqItems = document.querySelectorAll('.faq-accordion-item--modern');
    faqItems.forEach(item => {
        const trigger = item.querySelector('.faq-trigger');
        if (trigger) {
            trigger.addEventListener('click', function() {
                const isOpen = item.classList.contains('is-open');
                // Close all other items
                faqItems.forEach(otherItem => {
                    if (otherItem !== item) {
                        otherItem.classList.remove('is-open');
                    }
                });
            });
        }
    });

    // CTA Banner Parallax Effect
    const ctaBanner = document.querySelector('.cta-banner-card--modern');
    if (ctaBanner) {
        window.addEventListener('scroll', () => {
            const rect = ctaBanner.getBoundingClientRect();
            const scrollPercent = (window.innerHeight - rect.top) / (window.innerHeight + rect.height);
            if (scrollPercent > 0 && scrollPercent < 1) {
                ctaBanner.style.transform = `translateY(${(scrollPercent - 0.5) * 20}px)`;
            }
        });
    }

    // Scroll Reveal for Contact Page Elements
    const revealElements = document.querySelectorAll('.contact-hero-content, .contact-dossier, .contact-form-wrapper, .faq-accordion-item, .cta-banner-card');
    revealElements.forEach((element, index) => {
        element.style.opacity = '0';
        element.style.transform = 'translateY(30px)';
        element.style.transition = `opacity 0.6s ease ${index * 0.1}s, transform 0.6s ease ${index * 0.1}s`;
        
        const observer = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    element.style.opacity = '1';
                    element.style.transform = 'translateY(0)';
                    observer.unobserve(element);
                }
            });
        }, { threshold: 0.1 });
        
        observer.observe(element);
    });
}

// ===== اسکرول سکشن‌به‌سکشنِ صفحه درباره‌ما =====
// هر حرکت چرخ موس یا ترک‌پد — چه یک تیک، چه ده تیک — فقط یک سکشن جابه‌جا
// می‌کند. سکشن‌هایی که از صفحه بلندترند اول به اندازه یک صفحه خوانده می‌شوند
// و بعد سکشن بعدی می‌آیند. روی لمسی، اسنپ ملایم CSS (proximity) فعال می‌شود
// تا اسکرول طبیعی دست‌نخورده باقی بماند.
// توجه: listenerها یک‌بار روی window/document بسته می‌شوند و در هر رویداد
// عناصر را دوباره جستجو می‌کنند، پس با htmx swapها مشکلی پیش نمی‌آید.
function initAboutSectionSnap() {
    if (window.__aboutSnapBound) return;
    if (!document.querySelector('.about-page')) return;
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
    window.__aboutSnapBound = true;

    var SNAP_OFFSET = 80; // ارتفاع هدر شناور (هماهنگ با scroll-padding-top)
    var snapAnimating = false;
    var gestureActive = false;
    var gestureTimer = null;
    var wheelAccum = 0;

    function smoothScrollTo(targetY, duration) {
        var startY = window.scrollY || window.pageYOffset || 0;
        var distance = targetY - startY;
        if (Math.abs(distance) < 2) return;
        snapAnimating = true;
        var t0 = performance.now();

        function ease(t) {
            return t < 0.5 ? 4 * t * t * t : 1 - Math.pow(-2 * t + 2, 3) / 2;
        }

        function step(now) {
            var t = Math.min(1, (now - t0) / duration);
            window.scrollTo(0, startY + distance * ease(t));
            if (t < 1) requestAnimationFrame(step);
            else snapAnimating = false;
        }

        requestAnimationFrame(step);
    }

    function getSnapEls() {
        var page = document.querySelector('.about-page');
        if (!page) return [];
        // فقط hero + سکشن‌های شماره‌دار نقطه توقف بشن؛ manifesto جزء
        // نمای اولیه است (با hero یک نما محسوب می‌شود) و توقفِ جدا ندارد.
        return Array.prototype.slice.call(page.children).filter(function (el) {
            return (el.matches('header.hero, section')) &&
                   !el.classList.contains('manifesto');
        });
    }

    function nearestSnapIndex(els) {
        var best = 0;
        var bestDist = Infinity;
        for (var i = 0; i < els.length; i++) {
            var d = Math.abs(els[i].getBoundingClientRect().top - SNAP_OFFSET);
            if (d < bestDist) { bestDist = d; best = i; }
        }
        return best;
    }

    function goSnap(dir) {
        var els = getSnapEls();
        if (els.length < 2) return;
        var i = nearestSnapIndex(els);
        var el = els[i];
        var free = window.innerHeight - SNAP_OFFSET;

        // سکشن بلندتر از صفحه: اول بقیه‌اش را نشان بده، بعد برو بعدی
        if (el && el.offsetHeight > free + 8) {
            var rect = el.getBoundingClientRect();
            if (dir > 0 && rect.bottom > window.innerHeight + 8) {
                smoothScrollTo((window.scrollY || 0) + Math.min(free, rect.bottom - window.innerHeight + 4), 650);
                return;
            }
            if (dir < 0 && rect.top < SNAP_OFFSET - 8) {
                smoothScrollTo(rect.top + (window.scrollY || 0) - SNAP_OFFSET, 650);
                return;
            }
        }

        var next = i + dir;
        if (next < 0 || next >= els.length) return;
        smoothScrollTo(els[next].getBoundingClientRect().top + (window.scrollY || 0) - SNAP_OFFSET, 760);
    }

    function lockGesture() {
        gestureActive = true;
        clearTimeout(gestureTimer);
        gestureTimer = setTimeout(function () { gestureActive = false; }, 220);
    }

    // چرخ موس / ترک‌پد: تیک‌ها را جمع می‌کنیم تا هر رگبار حرکت = یک سکشن
    window.addEventListener('wheel', function (e) {
        if (e.ctrlKey) return; // پینچ‌زوم دست‌نخورده بماند
        if (!e.target || !e.target.closest) return;
        if (e.target.closest('.site-header, .mobile-nav, [data-no-snap]')) return;
        if (!document.querySelector('.about-page')) return;
        e.preventDefault();

        wheelAccum += e.deltaY;
        if (snapAnimating || gestureActive) { wheelAccum = 0; return; }
        if (Math.abs(wheelAccum) < 12) return;

        var dir = wheelAccum > 0 ? 1 : -1;
        wheelAccum = 0;
        lockGesture();
        goSnap(dir);
    }, { passive: false });

    // کیبورد: هر کلید = یک سکشن
    window.addEventListener('keydown', function (e) {
        if (!document.querySelector('.about-page')) return;
        var tag = (e.target && e.target.tagName) || '';
        if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' ||
            (e.target && e.target.isContentEditable)) return;

        var dir = 0;
        if (e.key === 'ArrowDown' || e.key === 'PageDown') dir = 1;
        else if (e.key === 'ArrowUp' || e.key === 'PageUp') dir = -1;
        else if (e.key === 'Home') {
            e.preventDefault();
            if (!snapAnimating) smoothScrollTo(0, 700);
            return;
        } else if (e.key === 'End') {
            e.preventDefault();
            var els = getSnapEls();
            if (!snapAnimating && els.length) {
                smoothScrollTo(els[els.length - 1].getBoundingClientRect().top + (window.scrollY || 0) - SNAP_OFFSET, 700);
            }
            return;
        } else {
            return;
        }

        e.preventDefault();
        if (e.repeat || snapAnimating || gestureActive) return;
        lockGesture();
        goSnap(dir);
    });

    // لمسی: اسنپ ملایم (proximity) — اسکرول طبیعی سرجایش می‌ماند
    document.documentElement.classList.add('about-snap');

    // لینک‌های لنگرِ داخل صفحه با همان سیستم اسکرول می‌شوند (delegation
    // است تا بعد از htmx swap هم روی لینک‌های تازه کار کند)
    document.addEventListener('click', function (e) {
        var a = e.target && e.target.closest &&
            e.target.closest('.about-page [data-anchor], .about-page a[href^="#"]');
        if (!a) return;
        var href = a.getAttribute('href') || '';
        if (href.length < 2 || href.charAt(0) !== '#') return;
        var target = document.querySelector(href);
        if (!target) return;
        e.preventDefault();
        if (snapAnimating) return;
        var startY = window.scrollY || window.pageYOffset || 0;
        var endY = target.getBoundingClientRect().top + startY - SNAP_OFFSET;
        smoothScrollTo(endY, Math.min(1400, Math.max(600, Math.abs(endY - startY) * 0.55)));
    });
}

// ===== صفحه‌بار (لودر) =====
// تضمین می‌کند لودر حتی با درخواست‌های سریع دیده شود (حداقل مدت نمایش ثابت)
(function () {
    var hideTimer = null;
    function show() {
        clearTimeout(hideTimer);
        document.body.classList.add('page-loader--visible');
    }
    function hideSoon() {
        clearTimeout(hideTimer);
        hideTimer = setTimeout(function () {
            document.body.classList.remove('page-loader--visible');
        }, 350);
    }
    document.addEventListener('htmx:beforeRequest', show);
    document.addEventListener('htmx:afterRequest', hideSoon);
    document.addEventListener('htmx:sendError', hideSoon);
})();

// ===== توست نوتیفیکیشن شیشه‌ای (فرم تماس) =====
function showToast(type, title, message) {
    var stack = document.getElementById('toast-stack');
    if (!stack) return;
    var toast = document.createElement('div');
    toast.setAttribute('role', 'status');
    toast.className = 'toast toast--' + type;

    var iconWrap = document.createElement('span');
    iconWrap.className = 'toast__icon';
    var icon = document.createElement('i');
    icon.setAttribute('data-lucide', type === 'success' ? 'check-circle-2' : 'alert-circle');
    iconWrap.appendChild(icon);

    var msgWrap = document.createElement('div');
    msgWrap.className = 'toast__msg';
    var strong = document.createElement('strong');
    strong.textContent = title;
    var small = document.createElement('span');
    small.textContent = message;
    msgWrap.appendChild(strong);
    msgWrap.appendChild(small);

    toast.appendChild(iconWrap);
    toast.appendChild(msgWrap);
    stack.appendChild(toast);
    if (window.lucide) lucide.createIcons();

    setTimeout(function () { toast.classList.add('toast--leave'); }, 3300);
    setTimeout(function () { if (toast.parentNode) toast.parentNode.removeChild(toast); }, 3800);
}

// با پایان swap نتیجه فرم تماس، توست موفق/ناموفق نمایش داده می‌شود
document.addEventListener('htmx:afterSwap', function (evt) {
    var target = evt.detail && evt.detail.target;
    if (!target || !target.matches || !target.matches('#contact-result')) return;
    var box = target.querySelector('.contact-response');
    if (!box) return;
    var ok = box.classList.contains('success');
    var titleEl = box.querySelector('strong');
    var textEl = box.querySelector('p');
    showToast(
        ok ? 'success' : 'error',
        ok ? 'پیام شما با موفقیت ثبت شد' : 'ارسال ناموفق بود',
        textEl ? textEl.textContent : (titleEl ? titleEl.textContent : 'لطفاً دوباره امتحان کنید')
    );
    target.innerHTML = '';
});

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
    initContactPageAnimations();
    initAboutCardSlider();
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
    initContactPageAnimations();
    initAboutCardSlider();
});
window.addEventListener('popstate', updateActiveNavigation);

// ===================================================================
//  اسلایدر کارتی تمام‌صفحه — صفحه درباره‌ما
//  هر سکشن یک «کارت» با ارتفاع دقیقاً 100% ارتفاع نما. به‌جای اسکرول،
//  هر حرکت چرخ موس/ترک‌پد (چه یک تیک، چه ده تیک) فقط یک کارت را با
//  انیمیشن بالا/پایین می‌بره. لمسی با swipe هندل می‌شود. listenerها
//  یک‌بار بسته می‌شن و عناصر هر گام دوباره جستجو می‌شن (برای htmx swapها).
// ===================================================================
var aboutSlider = {
    current: 0,
    animating: false,
    gesture: false,
    timer: null,
    unlockTimer: null,
    accum: 0
};
var aboutSliderBound = false;
var aboutMQBound = false;

function getAboutSlides() {
    var page = document.querySelector('.about-page');
    if (!page) return [];
    return Array.prototype.slice.call(page.children).filter(function (el) {
        return el.matches('header.hero, section');
    });
}

function layoutAboutSlides() {
    var slides = getAboutSlides();
    for (var i = 0; i < slides.length; i++) {
        slides[i].style.transform = 'translateY(' + (i - aboutSlider.current) * 100 + 'vh)';
    }
}

// حالت کارت تمام‌صفحه فقط وقتی معنا دارد که چیدمان دوستونه باشد
// (عرض ≥ 1101px) و ارتفاع پنجره برای یک نمای کامل کافی باشد (≥ 720px).
// زیر این مقادیر — موبایل و تبلت — سکشن‌ها تک‌ستونه و بلندتر از یک
// نما می‌شوند؛ کارتی آنجا هم محتوای زیرِ viewport را می‌بُرد و اسکرول
// لمسی را خراب می‌کند. پس صفحه به اسکرول عادی برمی‌گردد و فوتر دوباره
// دیده می‌شود (کلاس about-active هم گذاشته نمی‌شود).
var aboutCardMQ = (typeof window.matchMedia === 'function')
    ? window.matchMedia('(min-width: 1101px) and (min-height: 720px)')
    : { matches: true };

function aboutCardEnabled() {
    if (!aboutCardMQ.matches) return false;
    return !window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

// خروج از حالت کارتی (مثلاً وقتی پنجره را تا عرض موبایل کوچک می‌کنیم):
// ترنسفورم‌های inline که سکشن‌ها را جابه‌جا کرده‌اند پاک می‌شوند تا
// سکشن‌ها در جریان عادی سند برگردند و هیچ محتوایی بریده نماند.
function clearAboutSlideLayout() {
    var slides = getAboutSlides();
    for (var i = 0; i < slides.length; i++) {
        slides[i].style.transform = '';
    }
    aboutSlider.current = 0;
    aboutSlider.animating = false;
    clearTimeout(aboutSlider.unlockTimer);
}

function aboutSliderGo(dir) {
    var slides = getAboutSlides();
    if (slides.length < 2) return;
    var target = aboutSlider.current + dir;
    if (target < 0 || target >= slides.length) return;
    aboutSlider.animating = true;
    aboutSlider.current = target;
    layoutAboutSlides();
    setTimeout(function () { animateAboutCard(getAboutSlides()[target]); }, 160);
    clearTimeout(aboutSlider.unlockTimer);
    aboutSlider.unlockTimer = setTimeout(function () { aboutSlider.animating = false; }, 860);
}

function lockAboutGesture() {
    aboutSlider.gesture = true;
    clearTimeout(aboutSlider.timer);
    aboutSlider.timer = setTimeout(function () { aboutSlider.gesture = false; }, 240);
}

// Reveal ورود هر کارت: اجزاء [data-reveal] آن با تأخیر پله‌پله (stagger)
// ظاهر می‌شن. قبل از افشودن، حالت مخفی (is-armed) دوباره برقرار می‌شود
// تا با دوباره ورود به کارت، انیمیشن تکرار پذیرد.
function animateAboutCard(cardEl) {
    if (!cardEl) return;
    var els = Array.prototype.slice.call(cardEl.querySelectorAll('[data-reveal]'));
    els.forEach(function (el, i) {
        el.classList.remove('is-visible');
        el.classList.add('is-armed');
        el.style.transitionDelay = (80 + i * 90) + 'ms';
        void el.offsetHeight; // force reflow تا حالت مخفی اعمال شود
        requestAnimationFrame(function () {
            el.classList.add('is-visible');
        });
    });
}

function initAboutCardSlider() {
    // فوتر (که بعد از انرژی درباره در قالب است) نباید در این نما دیده شود؛
    // چون .about-page فیکس است، این کلاس روی body فوتر را در این صفحه مخفی می‌کند
    // (با htmx boost که فوتر را دوباره رندر نمیکند، این روش امن است).
    // فقط در حالت کارتی (دسکتاپ) گذاشته می‌شود؛ در موبایل فوتر باید برگردد.
    var hasPage = !!document.querySelector('.about-page');
    var cardMode = hasPage && aboutCardEnabled();
    document.body.classList.toggle('about-active', cardMode);

    // عبور از مرز عرض/ارتفاع (resize، چرخش گوشی) → این تابع دوباره صدا
    // زده می‌شود و حالت درست را انتخاب می‌کند. باید قبل از هر return
    // ثبت شود، وگرنه اگر صفحه اول روی موبایل باز شود هرگز ثبت نمی‌شود.
    if (!aboutMQBound && typeof aboutCardMQ.addEventListener === 'function') {
        aboutMQBound = true;
        aboutCardMQ.addEventListener('change', initAboutCardSlider);
    } else if (!aboutMQBound && typeof aboutCardMQ.addListener === 'function') {
        aboutMQBound = true;
        aboutCardMQ.addListener(initAboutCardSlider);
    }

    if (!hasPage) return;

    // موبایل / تبلت / پنجره‌ی کوتاه / حرکت محدود:
    // اسکرول عادی، بدون هیچ دخالتی
    if (!cardMode) {
        clearAboutSlideLayout();
        return;
    }

    // در هر ورود به صفحه از کارت اول شروع و موقعیت را به‌روز کن
    aboutSlider.current = 0;
    layoutAboutSlides();
    animateAboutCard(getAboutSlides()[0]);

    if (aboutSliderBound) return;
    aboutSliderBound = true;

    // چرخ موس / ترک‌پد: تیک‌ها جمع می‌شن؛ هر رگبار حرکت = یک کارت
    window.addEventListener('wheel', function (e) {
        if (!aboutCardEnabled()) return;
        if (e.ctrlKey) return; // پینچ‌زوم دست‌نخورده بماند
        if (!e.target || !e.target.closest) return;
        if (e.target.closest('.site-header, .mobile-nav, [data-no-snap]')) return;
        if (!document.querySelector('.about-page')) return;
        e.preventDefault();
        aboutSlider.accum += e.deltaY;
        if (aboutSlider.animating || aboutSlider.gesture) { aboutSlider.accum = 0; return; }
        if (Math.abs(aboutSlider.accum) < 12) return;
        var dir = aboutSlider.accum > 0 ? 1 : -1;
        aboutSlider.accum = 0;
        lockAboutGesture();
        aboutSliderGo(dir);
    }, { passive: false });

    // لمسی: swipe بالا/پایین = یک کارت (فقط در حالت کارتی؛
    // در موبایل اسکرول طبیعی نباید دستکاری شود)
    var tsY = null;
    window.addEventListener('touchstart', function (e) {
        if (!aboutCardEnabled()) { tsY = null; return; }
        if (e.target && e.target.closest && e.target.closest('.site-header, .mobile-nav')) { tsY = null; return; }
        tsY = e.touches[0].clientY;
    }, { passive: true });
    window.addEventListener('touchend', function (e) {
        if (tsY === null) return;
        if (!aboutCardEnabled()) { tsY = null; return; }
        var dy = tsY - (e.changedTouches[0].clientY || tsY);
        tsY = null;
        if (Math.abs(dy) < 40 || aboutSlider.animating || aboutSlider.gesture) return;
        aboutSliderGo(dy > 0 ? 1 : -1);
    }, { passive: true });

    // کیبورد: هر کلید = یک کارت
    window.addEventListener('keydown', function (e) {
        if (!aboutCardEnabled()) return;
        if (!document.querySelector('.about-page')) return;
        var tag = (e.target && e.target.tagName) || '';
        if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' ||
            (e.target && e.target.isContentEditable)) return;
        var dir = 0;
        if (e.key === 'ArrowDown' || e.key === 'PageDown') dir = 1;
        else if (e.key === 'ArrowUp' || e.key === 'PageUp') dir = -1;
        else if (e.key === 'Home') {
            e.preventDefault();
            if (!aboutSlider.animating) { aboutSlider.current = 0; layoutAboutSlides(); }
            return;
        } else if (e.key === 'End') {
            e.preventDefault();
            var els = getAboutSlides();
            if (!aboutSlider.animating && els.length) { aboutSlider.current = els.length - 1; layoutAboutSlides(); }
            return;
        } else {
            return;
        }
        e.preventDefault();
        if (e.repeat || aboutSlider.animating || aboutSlider.gesture) return;
        lockAboutGesture();
        aboutSliderGo(dir);
    });

    // لنگر داخل کارت (مثل #ssi): پریدن به همان کارت
    document.addEventListener('click', function (e) {
        if (!aboutCardEnabled()) return;
        var a = e.target && e.target.closest &&
            e.target.closest('.about-page [data-anchor], .about-page a[href^="#"]');
        if (!a) return;
        var href = a.getAttribute('href') || '';
        if (href.length < 2 || href.charAt(0) !== '#') return;
        var target = document.querySelector(href);
        if (!target) return;
        e.preventDefault();
        var slides = getAboutSlides();
        var idx = slides.indexOf(target);
        if (idx >= 0 && !aboutSlider.animating) {
            aboutSlider.current = idx;
            layoutAboutSlides();
        }
    });
}

