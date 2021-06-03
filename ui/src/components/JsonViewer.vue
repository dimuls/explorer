<style>
.el-drawer__body {
  position: absolute;
  top: 66.75px;
  bottom: 0;
  left: 0;
  right: 0;
  overflow: auto;
}
</style>

<template>
  <div class="tree" ref="tree"></div>
</template>

<script>
import JSONTreeView from "json-tree-view";
import "json-tree-view/devtools.css";

export default {
  name: "JsonViewer",
  props: {
    js: {},
    json: {
      type: String,
    },
    rootName: {
      type: String,
      default: "",
    },
  },
  mounted() {
    let js;
    if (this.js) {
      js = this.js;
    } else {
      js = JSON.parse(this.json);
    }
    const viewer = new JSONTreeView(this.rootName, js);
    viewer.expand(true);
    viewer.readonly = true;
    this.$refs.tree.appendChild(viewer.dom);
  },
};
</script>
