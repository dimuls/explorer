<template>
  <div ref="container">
    <slot></slot>
  </div>
</template>

<script>
import debounce from "lodash.debounce";

export default {
  name: "InfiniteScroll",
  emits: ["load-more"],
  props: {
    complete: {
      type: Boolean,
      default: false,
    },
  },
  watch: {
    complete: {
      immediate: true,
      handler(complete) {
        if (complete) {
          window.removeEventListener("scroll", this.handleScroll);
        } else {
          window.addEventListener("scroll", this.handleScroll);
        }
      },
    },
  },
  methods: {
    emitLoadMore: debounce(function () {
      this.$emit("load-more");
    }, 200),
    handleScroll() {
      if (
        this.$refs.container.getBoundingClientRect().bottom < window.innerHeight
      ) {
        this.emitLoadMore();
      }
    },
  },
  mounted() {
    window.addEventListener("scroll", this.handleScroll);
  },
  unmounted() {
    window.removeEventListener("scroll", this.handleScroll);
  },
};
</script>
