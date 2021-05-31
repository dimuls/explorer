import { createApp } from "vue";
import "element-plus/lib/theme-chalk/index.css";
import axios from "axios";
import VueAxios from "vue-axios";
import App from "./App";
import router from "./router";
import store from "./store";
import basicExplorer from "@basic/plugin";

createApp(App)
  .use(store)
  .use(router)
  .use(basicExplorer)
  .use(
    VueAxios,
    axios.create({
      baseURL: "http://localhost:8080/api/",
    })
  )
  .mount("#app");
