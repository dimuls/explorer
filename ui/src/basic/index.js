export { default as App } from "./App";
export { default as router } from "./router";
export { default as store } from "./store";

import query from "./mixins/query";

export const mixins = [query];
