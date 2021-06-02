import { createApp } from "vue";
import plugin from "./plugin";

import App from "./App";
import router from "./router";
import store from "./store";

createApp(App)
  .use(store)
  .use(router)
  .use(plugin, {
    apiBaseURL: "http://localhost:8080/api/",
  })
  .mount("#app");
