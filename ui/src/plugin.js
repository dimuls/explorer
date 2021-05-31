import ElementPlus from "element-plus";

import Blocks from "./views/Blocks";
import Chaincodes from "./views/Chaincodes";
import Channels from "./views/Channels";
import Peers from "./views/Peers";
import States from "./views/States";
import Transactions from "./views/Transactions";

import query from "./mixins/query";

const views = [Blocks, Chaincodes, Channels, Peers, States, Transactions];
const components = [];
const mixins = [query];

export default {
  install: (app) => {
    app.use(ElementPlus);
    views.forEach((v) => app.component(v.name, v));
    components.forEach((c) => app.component(c.name, c));
    mixins.forEach((m) => app.mixin(m));
  },
};
