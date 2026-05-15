import { createRouter, createWebHistory } from "vue-router";
import Scraper from "../views/Scraper.vue";
import Listings from "../views/Listings.vue";

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", redirect: "/listings" },
    { path: "/scraper", name: "scraper", component: Scraper },
    { path: "/listings", name: "listings", component: Listings },
    { path: "/:pathMatch(.*)*", redirect: "/listings" },
  ],
});
