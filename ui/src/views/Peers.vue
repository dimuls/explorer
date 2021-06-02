<style scoped></style>

<template>
  <el-form :inline="true">
    <el-form-item>
      <el-select v-model="channelId" placeholder="Channel ID" clearable>
        <el-option
          v-for="c in channels"
          :key="c.id"
          :label="c.name"
          :value="c.id"
        ></el-option>
      </el-select>
    </el-form-item>
  </el-form>
  <el-table :data="peers" stripe>
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
      channelId: undefined,
      peers: [],
    };
  },
  watch: {
    channelId(channelId) {
      this.setQuery("channelId", channelId);
      this.reloadPeers();
    },
  },
  async mounted() {
    const res = await this.$http.get("/channels");
    if (res.data.channels.length) {
      this.channels.push(...res.data.channels);
    }
    this.channelId = this.$route.query.channelId;
    await this.reloadPeers();
  },
  methods: {
    async reloadPeers() {
      const res = await this.$http.get(
        "/peers" +
          this.encodeQuery({
            channelId: this.channelId,
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
