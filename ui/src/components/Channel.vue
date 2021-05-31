<style scoped></style>

<template></template>

<script>
export default {
  name: "Channel",
  props: {
    channelID: {
      type: Number,
      required: true,
    },
  },
  data() {
    return {
      channel: null,
    };
  },
  async mounted() {
    const res = this.$http.get(
      "/channel" +
        this.encodeQuery({
          channel_id: this.channelID,
        })
    );
  },
};
</script>
