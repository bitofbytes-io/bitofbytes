/* Nocturne behaviour: the full-size screenshot viewer. Plain script, no framework.
   Markup contract: a root with [data-nc-shots] holding figure.nc-shot, each with an <a href="full image"><img></a>
   and an optional <figcaption><b>title</b><span>note</span></figcaption>. Without this script (or without
   <dialog> support) the links open the image on its own, so nothing here is required to see a screenshot.
   The native modal <dialog> supplies the focus trap, Escape to close and focus return to the clicked link. */
(function () {
  'use strict';

  var roots = document.querySelectorAll('[data-nc-shots]');
  if (!roots.length || typeof HTMLDialogElement !== 'function') return;

  var dialog = document.createElement('dialog');
  if (typeof dialog.showModal !== 'function') return;

  var CHEVRON_LEFT = 'M15 5l-7 7 7 7';
  var CHEVRON_RIGHT = 'M9 5l7 7-7 7';
  var CLOSE = 'M6 6l12 12M18 6L6 18';

  function icon(path) {
    var svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    svg.setAttribute('viewBox', '0 0 24 24');
    svg.setAttribute('width', '22');
    svg.setAttribute('height', '22');
    svg.setAttribute('fill', 'none');
    svg.setAttribute('stroke', 'currentColor');
    svg.setAttribute('stroke-width', '2');
    svg.setAttribute('stroke-linecap', 'round');
    svg.setAttribute('stroke-linejoin', 'round');
    svg.setAttribute('aria-hidden', 'true');
    var p = document.createElementNS('http://www.w3.org/2000/svg', 'path');
    p.setAttribute('d', path);
    svg.appendChild(p);
    return svg;
  }

  function button(className, label, path) {
    var b = document.createElement('button');
    b.type = 'button';
    b.className = 'nc-viewer__btn ' + className;
    b.setAttribute('aria-label', label);
    b.appendChild(icon(path));
    return b;
  }

  function el(tag, className) {
    var node = document.createElement(tag);
    node.className = className;
    return node;
  }

  dialog.className = 'nc-viewer';
  dialog.setAttribute('aria-label', 'Screenshot viewer');

  var closeBtn = button('nc-viewer__close', 'Close', CLOSE);
  closeBtn.autofocus = true;
  var prevBtn = button('nc-viewer__prev', 'Previous screenshot', CHEVRON_LEFT);
  var nextBtn = button('nc-viewer__next', 'Next screenshot', CHEVRON_RIGHT);
  var stage = el('div', 'nc-viewer__stage');
  var image = el('img', 'nc-viewer__img');
  var caption = el('p', 'nc-viewer__caption');
  var title = el('b', 'nc-viewer__title');
  var note = el('span', 'nc-viewer__note');
  var count = el('span', 'nc-viewer__count');
  count.setAttribute('aria-live', 'polite');

  caption.appendChild(title);
  caption.appendChild(note);
  caption.appendChild(count);
  stage.appendChild(prevBtn);
  stage.appendChild(image);
  stage.appendChild(nextBtn);
  dialog.appendChild(closeBtn);
  dialog.appendChild(stage);
  dialog.appendChild(caption);
  document.body.appendChild(dialog);

  var items = [];
  var index = 0;

  function show(i) {
    index = (i + items.length) % items.length;
    var item = items[index];
    image.src = item.src;
    image.alt = item.alt;
    title.textContent = item.title;
    note.textContent = item.note;
    count.textContent = (index + 1) + ' of ' + items.length;
  }

  function open(root, link) {
    items = Array.prototype.map.call(root.querySelectorAll('figure.nc-shot a'), function (a) {
      var img = a.querySelector('img');
      var figure = a.closest('figure');
      var t = figure && figure.querySelector('figcaption b');
      var n = figure && figure.querySelector('figcaption span');
      return {
        link: a,
        src: a.href,
        alt: img ? img.alt : '',
        title: t ? t.textContent : '',
        note: n ? n.textContent : ''
      };
    });
    var start = items.findIndex(function (item) { return item.link === link; });
    if (start < 0) return false;
    var several = items.length > 1;
    prevBtn.hidden = nextBtn.hidden = count.hidden = !several;
    show(start);
    dialog.showModal();
    return true;
  }

  function step(delta) {
    if (items.length > 1) show(index + delta);
  }

  Array.prototype.forEach.call(roots, function (root) {
    root.addEventListener('click', function (event) {
      var link = event.target.closest && event.target.closest('a');
      if (!link || !root.contains(link) || !link.querySelector('img')) return;
      // Let the browser handle "open in new tab/window" and download gestures.
      if (event.defaultPrevented || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
      if (open(root, link)) event.preventDefault();
    });
  });

  prevBtn.addEventListener('click', function () { step(-1); });
  nextBtn.addEventListener('click', function () { step(1); });
  closeBtn.addEventListener('click', function () { dialog.close(); });

  dialog.addEventListener('keydown', function (event) {
    if (event.key === 'ArrowLeft') {
      step(-1);
      event.preventDefault();
    } else if (event.key === 'ArrowRight') {
      step(1);
      event.preventDefault();
    }
  });

  // A click on the backdrop lands on the dialog itself, or on the empty stage around the image.
  dialog.addEventListener('click', function (event) {
    if (event.target === dialog || event.target === stage) dialog.close();
  });

  dialog.addEventListener('close', function () {
    image.removeAttribute('src');
  });
}());
