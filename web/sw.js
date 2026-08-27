/* SubGame service worker — shell + wasm cache for installed / offline-friendly play. */
const CACHE = "subgame-v2";
const PRECACHE = [
  "./",
  "./index.html",
  "./manifest.json",
  "./wasm_exec.js",
  "./icons/icon-192.png",
  "./icons/icon-512.png",
  "./icons/icon-180.png",
];

self.addEventListener("install", (event) => {
  event.waitUntil(
    caches.open(CACHE).then((cache) => cache.addAll(PRECACHE)).then(() => self.skipWaiting())
  );
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches.keys().then((keys) =>
      Promise.all(keys.filter((k) => k !== CACHE).map((k) => caches.delete(k)))
    ).then(() => self.clients.claim())
  );
});

function isSameOrigin(url) {
  return url.origin === self.location.origin;
}

function isWasmPath(pathname) {
  return pathname.endsWith("/game.wasm") || pathname.endsWith("game.wasm");
}

self.addEventListener("fetch", (event) => {
  const req = event.request;
  if (req.method !== "GET") return;

  const url = new URL(req.url);
  if (!isSameOrigin(url)) return;

  if (isWasmPath(url.pathname)) {
    event.respondWith(
      caches.open(CACHE).then(async (cache) => {
        const cached = await cache.match("./game.wasm");
        if (cached) return cached;
        const res = await fetch(req);
        if (res.ok) {
          cache.put("./game.wasm", res.clone());
        }
        return res;
      })
    );
    return;
  }

  event.respondWith(
    caches.match(req).then((cached) => {
      if (cached) return cached;
      return fetch(req).then((res) => {
        const cacheable =
          res.ok &&
          (url.pathname.endsWith(".js") ||
            url.pathname.endsWith(".html") ||
            url.pathname.endsWith("/") ||
            url.pathname.endsWith(".webmanifest") ||
            url.pathname.endsWith(".png"));
        if (cacheable) {
          const copy = res.clone();
          caches.open(CACHE).then((cache) => cache.put(req, copy));
        }
        return res;
      });
    })
  );
});
