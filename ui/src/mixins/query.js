import debounce from "lodash.debounce";

export default {
  data() {
    return {
      queryChanges: [],
    };
  },
  methods: {
    encodeQuery(query) {
      const q = new URLSearchParams();
      Object.keys(query).forEach((k) => {
        if (query[k]) {
          q.set(k, query[k]);
        }
      });
      let qq = q.toString();
      if (qq.length > 0) {
        qq = "?" + qq;
      }
      return qq;
    },
    setQuery(name, value) {
      console.log(name, value);
      this.queryChanges.push({ [name]: value || undefined });
      this.__updateQuery();
    },
    __updateQuery: debounce(function () {
      const query = this.queryChanges.reduce(
        (qc, q) => ({ ...qc, ...q }),
        this.$route.query
      );
      this.$router.push({ query });
    }, 100),
  },
};
