export default {
  methods: {
    encodeQuery(query) {
      const q = new URLSearchParams();
      Object.keys(query).forEach((k) => {
        q.set(k, query[k]);
      });
      if (q.entries().length > 0) {
        return "?" + q.toString();
      }
      return "";
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
