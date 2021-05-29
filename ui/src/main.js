import { createApp } from "vue";
import store from "./store";
import ElementPlus from "element-plus";
import "element-plus/lib/theme-chalk/index.css";

const mod = require("./mods/" + (process.env.VUE_APP_MOD || "basic"));

const app = createApp(mod.app);

app.use(store);
app.use(mod.router);
app.use(ElementPlus);

app.mount("#app");
