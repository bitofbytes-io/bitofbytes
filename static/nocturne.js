/* Nocturne behaviour: the octave keyboard sound. Plain script, no framework.
   Exposes window.Nocturne = { NOTES, playNote, attachOctave, autoAttach }.
   Markup contract: a root with [data-nc-octave], keys carrying data-note ("C4".."C5", "C#4".."A#4"),
   and optionally a <button data-nc-sound aria-pressed="true"> anywhere on the page. */
(function () {
  'use strict';

  var NOTES = {
    'C4': 261.63, 'C#4': 277.18, 'D4': 293.66, 'D#4': 311.13, 'E4': 329.63, 'F4': 349.23,
    'F#4': 369.99, 'G4': 392.0, 'G#4': 415.3, 'A4': 440.0, 'A#4': 466.16, 'B4': 493.88, 'C5': 523.25
  };
  var WHITE_ORDER = ['C4', 'D4', 'E4', 'F4', 'G4', 'A4', 'B4', 'C5'];
  var STORAGE_KEY = 'nocturne-sound';

  var ctx = null;
  var soundOn = true;
  try { soundOn = window.localStorage.getItem(STORAGE_KEY) !== 'off'; } catch (e) { soundOn = true; }

  function audio() {
    if (!ctx) {
      var AC = window.AudioContext || window.webkitAudioContext;
      if (!AC) return null;
      ctx = new AC();
    }
    if (ctx.state === 'suspended') {
      try { ctx.resume().catch(function () {}); } catch (e) {}
    }
    return ctx;
  }

  // A soft, short piano-like pluck: triangle fundamental plus two quiet sine partials through a closing low-pass.
  // Returns false when nothing sounded because audio is still locked (see unlockAndPlay).
  function playNote(note) {
    if (!soundOn) return true;
    var freq = typeof note === 'number' ? note : NOTES[note];
    if (!freq) return true;
    var c = ctx; // only a user gesture creates or resumes the context
    if (!c || c.state !== 'running') return false;
    var t = c.currentTime;
    var out = c.createGain();
    out.gain.setValueAtTime(0.0001, t);
    out.gain.exponentialRampToValueAtTime(0.16, t + 0.012);
    out.gain.exponentialRampToValueAtTime(0.0001, t + 1.6);
    var lp = c.createBiquadFilter();
    lp.type = 'lowpass';
    lp.frequency.setValueAtTime(freq * 6, t);
    lp.frequency.exponentialRampToValueAtTime(freq * 1.5, t + 1.2);
    lp.connect(out);
    out.connect(c.destination);
    [[1, 'triangle', 1], [2, 'sine', 0.35], [3, 'sine', 0.12]].forEach(function (p) {
      var osc = c.createOscillator();
      osc.type = p[1];
      osc.frequency.value = freq * p[0];
      var g = c.createGain();
      g.gain.value = p[2];
      osc.connect(g);
      g.connect(lp);
      osc.start(t);
      osc.stop(t + 1.7);
    });
    return true;
  }

  // Call only while handling a user gesture (click, tap, key press): starts or resumes audio, then plays.
  function unlockAndPlay(note) {
    if (!soundOn) return;
    var c = audio();
    if (!c) return;
    if (c.state === 'running') { playNote(note); return; }
    try { c.resume().then(function () { playNote(note); }).catch(function () {}); } catch (e) {}
  }

  function setSound(on, fromGesture) {
    soundOn = on;
    try { window.localStorage.setItem(STORAGE_KEY, on ? 'on' : 'off'); } catch (e) {}
    var buttons = document.querySelectorAll('[data-nc-sound]');
    for (var i = 0; i < buttons.length; i++) {
      buttons[i].setAttribute('aria-pressed', on ? 'true' : 'false');
      var label = buttons[i].querySelector('[data-nc-sound-label]');
      if (label) label.textContent = on ? 'Sound on' : 'Sound off';
    }
    if (on && fromGesture) audio();
  }

  function attachOctave(root) {
    if (!root || root.__ncOctave) return;
    root.__ncOctave = true;
    var keys = root.querySelectorAll('.nc-key, .nc-black');
    for (var i = 0; i < keys.length; i++) {
      (function (el, index) {
        var note = el.getAttribute('data-note') || (el.classList.contains('nc-key') ? WHITE_ORDER[index] : null);
        var pending = false;
        el.addEventListener('pointerenter', function () {
          el.classList.add('is-pressed');
          // A touch enters before its gesture counts, so the first tap waits for pointerdown/up below.
          pending = !playNote(note);
        });
        function settle() {
          if (pending) { pending = false; unlockAndPlay(note); }
        }
        el.addEventListener('pointerdown', settle);
        el.addEventListener('pointerup', settle);
        el.addEventListener('pointerleave', function () { pending = false; el.classList.remove('is-pressed'); });
        el.addEventListener('pointercancel', function () { pending = false; el.classList.remove('is-pressed'); });
        // Keyboard users hear a white key when they tab onto it (the Tab press is the gesture).
        el.addEventListener('focus', function () {
          if (!el.matches(':focus-visible')) return;
          el.classList.add('is-pressed');
          unlockAndPlay(note);
        });
        el.addEventListener('blur', function () { el.classList.remove('is-pressed'); });
      })(keys[i], i);
    }
  }

  function autoAttach() {
    var roots = document.querySelectorAll('[data-nc-octave]');
    for (var i = 0; i < roots.length; i++) attachOctave(roots[i]);
    var buttons = document.querySelectorAll('[data-nc-sound]');
    for (var j = 0; j < buttons.length; j++) {
      if (buttons[j].__ncSound) continue;
      buttons[j].__ncSound = true;
      buttons[j].addEventListener('click', function () { setSound(!soundOn, true); });
    }
    setSound(soundOn, false);
    // Browsers keep audio locked until the first gesture; unlock on the first click, tap or key press anywhere.
    function unlock() {
      if (!soundOn) return;
      var c = audio();
      if (c && c.state === 'running') {
        document.removeEventListener('pointerdown', unlock);
        document.removeEventListener('keydown', unlock);
      }
    }
    document.addEventListener('pointerdown', unlock);
    document.addEventListener('keydown', unlock);
  }

  window.Nocturne = { NOTES: NOTES, playNote: playNote, attachOctave: attachOctave, autoAttach: autoAttach, setSound: setSound };

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', autoAttach);
  } else {
    autoAttach();
  }
})();
