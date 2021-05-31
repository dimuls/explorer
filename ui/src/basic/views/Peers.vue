<style scoped></style>

<template>
  <el-form :inline="true">
    <el-form-item label="">
      <el-select v-model="channelID" placeholder="Channel ID" clearable>
        <el-option
          v-for="c in channels"
          :key="c.id"
          :label="c.id"
          :value="c.id"
        ></el-option>
      </el-select>
    </el-form-item>
  </el-form>
  <el-table :data="peers">
    <el-table-column prop="id" label="ID"> </el-table-column>
    <el-table-column prop="url" label="URL"></el-table-column>
  </el-table>
</template>

<script>
export default {
  name: "Peers",
  data() {
    return {
      channels: [],
      channelID: "",
      peers: [],
    };
  },
  watch: {
    channelID(channelID) {
      this.setQuery("channel_id", channelID);
      this.reload();
    },
  },
  async mounted() {
    const res = await this.$http.get("/channels");
    if (res.data.channels) {
      this.channels.push(...res.data.channels);
    }
    this.reload();
  },
  methods: {
    async reload() {
      const res = await this.$http.get(
        "/peers" +
          this.encodeQuery({
            channel_id: this.channelID,
          })
      );

      this.peers.splice(0, this.peers.length);
      if (res.data.peers) {
        this.peers.push(...res.data.peers);
      }
    },
  },
};
</script>
