import { createApp } from "vue";
import ElementPlus from "element-plus";
import "element-plus/lib/theme-chalk/index.css";
import axios from "axios";
import VueAxios from "vue-axios";

const mod = require("./" + process.env.VUE_APP_MOD || "basic");

const app = createApp(mod.App)
  .use(mod.store)
  .use(mod.router)
  .use(ElementPlus)
  .use(
    VueAxios,
    axios.create({
      baseURL: "http://localhost:8080/api/",
    })
  );

mod.mixins.forEach((m) => {
  app.mixin(m);
});

app.mount("#app");
