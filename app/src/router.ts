import Vue from 'vue';
import Router from 'vue-router';

import Api from './Api';

import components from './components';

Vue.use(Router);

let router = new Router({
    routes: [
        { path: '/', component: components.albumsComponent, meta: { requiresAuth: true } },
        { path: '/albums', component: components.albumsComponent, meta: { requiresAuth: true } },
        { path: '/albums/:id', component: components.albumComponent, meta: { requiresAuth: true } },
        { path: '/current-queue', component: components.currentQueueComponent, meta: { requiresAuth: true } },
        { path: '/login', component: components.loginComponent },
        { path: '/playlists/:id', component: components.playlistComponent, meta: { requiresAuth: true } },
        { path: '/upload', component: components.uploadComponent, meta: { requiresAuth: true } },
    ],
});

router.beforeEach(function (to, from, next) {
    if (to.meta.requiresAuth && !Api.isAuthenticated()) {
        next('/login');
        return;
    }

    if (to.path === '/login' && Api.isAuthenticated()) {
        next('/');
        return;
    }

    return next();
});

export default router;
