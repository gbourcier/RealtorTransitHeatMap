import { createRouter, createWebHistory } from "vue-router";
import Scraper from "../views/Scraper.vue";
import Listings from "../views/Listings.vue";

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/scraper", name: "scraper", component: Scraper },
    { path: "/listings", name: "listings", component: Listings },
  ],
});
