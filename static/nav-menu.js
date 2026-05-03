document.addEventListener("pointerdown", function (event) {
  document.querySelectorAll(".nav-menu[open]").forEach(function (menu) {
    if (!menu.contains(event.target)) {
      menu.removeAttribute("open");
    }
  });
});

document.addEventListener("keydown", function (event) {
  if (event.key === "Escape") {
    document.querySelectorAll(".nav-menu[open]").forEach(function (menu) {
      menu.removeAttribute("open");
    });
  }
});
