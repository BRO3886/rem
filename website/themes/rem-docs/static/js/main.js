// Theme toggle
(function () {
  const toggle = document.getElementById('theme-toggle');
  const html = document.documentElement;

  const saved = localStorage.getItem('theme');
  if (saved) {
    html.setAttribute('data-theme', saved);
  } else if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
    html.setAttribute('data-theme', 'dark');
  }

  if (toggle) {
    toggle.addEventListener('click', function () {
      const current = html.getAttribute('data-theme');
      const next = current === 'dark' ? 'light' : 'dark';
      html.setAttribute('data-theme', next);
      localStorage.setItem('theme', next);
    });
  }
})();

// Mobile nav toggle
(function () {
  const btn = document.getElementById('nav-toggle');
  const links = document.getElementById('nav-links');
  if (btn && links) {
    btn.addEventListener('click', function () {
      links.classList.toggle('active');
    });
  }
})();

// Scroll reveal
(function () {
  var elements = document.querySelectorAll('.scroll-reveal');
  if (!elements.length) return;

  var observer = new IntersectionObserver(
    function (entries) {
      entries.forEach(function (entry) {
        if (entry.isIntersecting) {
          entry.target.classList.add('visible');
          observer.unobserve(entry.target);
        }
      });
    },
    { threshold: 0.12, rootMargin: '0px 0px -40px 0px' }
  );

  elements.forEach(function (el, i) {
    el.style.transitionDelay = (i % 3) * 0.08 + 's';
    observer.observe(el);
  });
})();

// Copy buttons
(function () {
  document.querySelectorAll('.copy-btn').forEach(function (btn) {
    btn.addEventListener('click', function () {
      var text = this.getAttribute('data-copy');
      navigator.clipboard.writeText(text).then(function () {
        btn.innerHTML =
          '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 6L9 17l-5-5"/></svg>';
        setTimeout(function () {
          btn.innerHTML =
            '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>';
        }, 1500);
      });
    });
  });
})();

// Install tabs
(function () {
  var tabs = document.querySelectorAll('.install-tab');
  var panels = document.querySelectorAll('.install-panel');
  if (!tabs.length) return;

  tabs.forEach(function (tab) {
    tab.addEventListener('click', function () {
      var target = this.getAttribute('data-tab');
      tabs.forEach(function (t) {
        t.classList.remove('active');
        t.setAttribute('aria-selected', 'false');
      });
      panels.forEach(function (p) { p.classList.remove('active'); });
      this.classList.add('active');
      this.setAttribute('aria-selected', 'true');
      var panel = document.querySelector('[data-panel="' + target + '"]');
      if (panel) panel.classList.add('active');
    });
  });
})();

// Terminal typing animation
(function () {
  var cmdEl = document.getElementById('typed-cmd');
  var outputEl = document.getElementById('terminal-output');
  if (!cmdEl || !outputEl) return;

  var demos = [
    {
      cmd: 'rem add "Ship v2.0" --due friday --priority high',
      output: 'Created reminder: Ship v2.0 (6ECEA745)',
    },
    {
      cmd: 'rem list --incomplete',
      output:
        'ID        List     Name             Due         Priority\n6ECEA745  Work     Ship v2.0        Fri Feb 13  High\nE24435D3  Work     Review PRs       Thu Feb 12  Medium\nA1B2C3D4  Home     Buy groceries    Today       None',
    },
    {
      cmd: 'rem complete 6ECE',
      output: 'Completed: Ship v2.0',
    },
    {
      cmd: 'rem stats',
      output:
        'Total: 224  Completed: 189  Incomplete: 35\nOverdue: 2  Flagged: 5  Completion: 84%',
    },
  ];

  var demoIdx = 0;
  var charIdx = 0;
  var typing = true;
  var pauseAfter = 2200;

  function typeNext() {
    var demo = demos[demoIdx];

    if (typing) {
      if (charIdx <= demo.cmd.length) {
        cmdEl.textContent = demo.cmd.slice(0, charIdx);
        charIdx++;
        setTimeout(typeNext, 32 + Math.random() * 28);
      } else {
        typing = false;
        setTimeout(function () {
          outputEl.textContent = demo.output;
          setTimeout(typeNext, pauseAfter);
        }, 300);
      }
    } else {
      outputEl.textContent = '';
      cmdEl.textContent = '';
      charIdx = 0;
      typing = true;
      demoIdx = (demoIdx + 1) % demos.length;
      setTimeout(typeNext, 400);
    }
  }

  setTimeout(typeNext, 800);
})();
