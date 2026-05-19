import { createRouter, createWebHistory } from "vue-router";
import Scraper from "../views/Scraper.vue";
import Listings from "../views/Listings.vue";
import Transit from "../views/Transit.vue";

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", redirect: "/listings" },
    { path: "/scraper", name: "scraper", component: Scraper, meta: { title: "Realtor Scraper", settings: true } },
    { path: "/transit", name: "transit", component: Transit, meta: { title: "GTFS Data", settings: true } },
    { path: "/listings", name: "listings", component: Listings, meta: { title: "Listings" } },
    { path: "/:pathMatch(.*)*", redirect: "/listings" },
  ],
});
