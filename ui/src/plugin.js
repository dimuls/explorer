import ElementPlus from "element-plus";
import "element-plus/lib/theme-chalk/index.css";
import axios from "axios";
import VueAxios from "vue-axios";

import Blocks from "./views/Blocks";
import Chaincodes from "./views/Chaincodes";
import Channels from "./views/Channels";
import Peers from "./views/Peers";
import States from "./views/States";
import Transactions from "./views/Transactions";

import InfiniteScroll from "./components/InfiniteScroll";
import JsonViewer from "./components/JsonViewer";

import queryMx from "@/mixins/query";
import dateMx from "@/mixins/date";

const views = [
  InfiniteScroll,
  JsonViewer,
  Blocks,
  Chaincodes,
  Channels,
  Peers,
  States,
  Transactions,
];
const components = [];
const mixins = [queryMx, dateMx];

export default {
  install: (app, options) => {
    app.use(ElementPlus).use(
      VueAxios,
      axios.create({
        baseURL: options.apiBaseURL,
      })
    );
    views.forEach((v) => app.component(v.name, v));
    components.forEach((c) => app.component(c.name, c));
    mixins.forEach((m) => app.mixin(m));
  },
};
