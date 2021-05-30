import { createRouter, createWebHistory } from "vue-router";
import { routes as basicRoutes } from "@/basic/router";
import Assets from "../views/Assets.vue";

export const routes = [
  { path: "/", name: "index", redirect: "assets" },
  { path: "/assets", name: "assets", component: Assets },
  ...basicRoutes.filter((r) => r.name !== "index"),
];

export default createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes,
});
