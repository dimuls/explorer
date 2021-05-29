import { createRouter, createWebHistory } from "vue-router";
import BasicApp from "../BasicApp";
import Peers from "../views/Peers";
import Channels from "../views/Channels";
import Chaincodes from "../views/Chaincodes";
import Blocks from "../views/Blocks";
import Transactions from "../views/Transactions";
import States from "../views/States";

export const app = BasicApp;

export const routes = [
  {
    path: "/",
    name: "index",
    redirect: "states",
  },
  {
    path: "/peers",
    name: "peers",
    component: Peers,
  },
  {
    path: "/channels",
    name: "channels",
    component: Channels,
  },
  {
    path: "/chaincodes",
    name: "chaincodes",
    component: Chaincodes,
  },
  {
    path: "/blocks",
    name: "blocks",
    component: Blocks,
  },
  {
    path: "/transactions",
    name: "transactions",
    component: Transactions,
  },
  {
    path: "/states",
    name: "states",
    component: States,
  },
];

export const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes,
});
