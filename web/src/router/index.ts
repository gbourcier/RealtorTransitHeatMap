import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useSavedFiltersStore } from '../stores/savedFilters'

const Login = () => import('../views/Login.vue')
const Listings = () => import('../views/Listings.vue')
const Favorites = () => import('../views/Favorites.vue')
const Scraper = () => import('../views/Scraper.vue')
const Transit = () => import('../views/Transit.vue')
const Users = () => import('../views/Users.vue')

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: Login, meta: { public: true } },
    { path: '/', redirect: '/listings' },
    { path: '/listings', name: 'listings', component: Listings, meta: { title: 'Listings' } },
    { path: '/favorites', name: 'favorites', component: Favorites, meta: { title: 'Manage Favorites' } },
    { path: '/scraper', name: 'scraper', component: Scraper, meta: { title: 'Realtor Scraper', settings: true, admin: true } },
    { path: '/transit', name: 'transit', component: Transit, meta: { title: 'GTFS Data', settings: true, admin: true } },
    { path: '/users', name: 'users', component: Users, meta: { title: 'Users', settings: true, admin: true } },
    { path: '/:pathMatch(.*)*', redirect: '/listings' },
  ],
})

router.beforeEach(async (to) => {
  const authStore = useAuthStore()
  if (!authStore.initialized) {
    await authStore.fetchMe()
  }
  if (to.meta.public) {
    if (to.path === '/login' && authStore.isLoggedIn) return '/listings'
    return
  }
  if (!authStore.isLoggedIn) return '/login'
  const savedFiltersStore = useSavedFiltersStore()
  if (!savedFiltersStore.initialized) {
    await savedFiltersStore.fetchList()
  }
  if (to.meta.admin && !authStore.isAdmin) return '/listings'
})
