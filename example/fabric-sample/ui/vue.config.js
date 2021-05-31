const path = require("path");

module.exports = {
  configureWebpack: {
    resolve: {
      alias: {
        vue: path.resolve("./node_modules/vue"),
        "@basic": path.resolve(__dirname, "../../../ui/src"),
      },
      extensions: [".vue", ".js", ".json"],
    },
  },
};
