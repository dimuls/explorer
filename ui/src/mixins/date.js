import dayjs from "dayjs";

export default {
  methods: {
    parseDate(str) {
      if (!str) {
        return undefined;
      }
      return new Date(str);
    },
    formatDate(date) {
      if (!date) {
        return "";
      }
      return dayjs(date).format("YYYY-MM-DDTHH:mm:ssZ");
    },
    readableDate(date) {
      return dayjs(date).format("YYYY-MM-DD HH:mm:ss");
    },
  },
};
