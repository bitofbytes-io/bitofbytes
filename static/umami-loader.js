(function () {
  if (window.location.hostname !== "bitofbytes.io") {
    return;
  }

  var script = document.createElement("script");
  script.async = true;
  script.dataset.websiteId = "16dc0a19-c8ad-4ede-a122-c0986d542941";
  script.src = "https://bitofbytes.io/umami/script.js";
  document.head.appendChild(script);
})();
