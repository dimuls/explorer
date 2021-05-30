import { createApp } from "vue";
import ElementPlus from "element-plus";
import "element-plus/lib/theme-chalk/index.css";

const mod = require("./" + process.env.VUE_APP_MOD || "basic");

createApp(mod.App)
  .use(mod.store)
  .use(mod.router)
  .use(ElementPlus)
  .mount("#app");
