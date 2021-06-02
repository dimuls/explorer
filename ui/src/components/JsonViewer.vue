<template>
  <div ref="container"></div>
</template>

<script>
import JSONTreeView from "json-tree-view";
import "json-tree-view/devtools.css";

export default {
  name: "JsonViewer",
  props: {
    json: {
      type: String,
    },
    rootName: {
      type: String,
      default: "",
    },
  },
  mounted() {
    const viewer = new JSONTreeView(this.rootName, JSON.parse(this.json));
    viewer.expand(true);
    viewer.readonly = true;
    this.$refs.container.appendChild(viewer.dom);
  },
};
</script>
