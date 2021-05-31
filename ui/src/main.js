import { createApp } from "vue";
import ElementPlus from "element-plus";
import "element-plus/lib/theme-chalk/index.css";
import axios from "axios";
import VueAxios from "vue-axios";
import App from "./App";
import router from "./router";
import store from "./store";
import queryMx from "./mixins/query";

createApp(App)
  .use(store)
  .use(router)
  .use(ElementPlus)
  .use(
    VueAxios,
    axios.create({
      baseURL: "http://localhost:8080/api/",
    })
  )
  .mixin(queryMx)
  .mount("#app");
