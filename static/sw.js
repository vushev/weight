const CACHE_NAME = 'weight-challenge-v1';
const urlsToCache = [
    '/',
    '/static/index.html',
    '/static/css/styles.css',
    '/static/js/config.js',
    '/static/js/utils.js',
    '/static/js/init.js',
    '/static/js/auth.js',
    '/static/js/weight.js',
    '/static/js/social.js',
    '/static/js/settings.js',
    '/static/js/challenges.js',
    '/static/js/calories.js',
    '/static/js/api.js'
];

self.addEventListener('install', event => {
    event.waitUntil(
        caches.open(CACHE_NAME)
            .then(cache => {
                const cachePromises = urlsToCache.map(url => {
                    return cache.add(url).catch(error => {
                        console.error('Failed to cache:', url, error);
                        return Promise.resolve();
                    });
                });
                return Promise.all(cachePromises);
            })
    );
});

self.addEventListener('fetch', event => {
    event.respondWith(
        caches.match(event.request)
            .then(response => {
                if (response) {
                    return response;
                }

                return fetch(event.request).then(response => {
                    if (!response || response.status !== 200 || response.type !== 'basic') {
                        return response;
                    }

                    const responseToCache = response.clone();
                    caches.open(CACHE_NAME)
                        .then(cache => {
                            cache.put(event.request, responseToCache);
                        });

                    return response;
                });
            })
            .catch(error => {
                console.error('Fetch error:', error);
                return new Response('Offline');
            })
    );
}); 