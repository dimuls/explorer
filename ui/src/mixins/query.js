export default {
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
      this.$router.push({
        query: {
          ...this.$route.query,
          [name]: value || undefined,
        },
      });
    },
  },
};
