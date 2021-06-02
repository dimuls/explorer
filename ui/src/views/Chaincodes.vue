<template>
  <el-form :inline="true">
    <el-form-item>
      <el-form-item>
        <el-select v-model="peerId" placeholder="Peer ID" clearable>
          <el-option
            v-for="c in peers"
            :key="c.id"
            :label="c.url"
            :value="c.id"
          ></el-option>
        </el-select>
      </el-form-item>
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
  <el-table :data="chaincodes" stripe>
    <el-table-column prop="id" label="ID"> </el-table-column>
    <el-table-column prop="name" label="Name"></el-table-column>
    <el-table-column prop="version" label="Version"></el-table-column>
  </el-table>
</template>

<script>
export default {
  name: "Chaincodes",
  data() {
    return {
      peers: [],
      peerId: undefined,
      channels: [],
      channelId: this.$route.query.channelId,
      chaincodes: [],
    };
  },
  watch: {
    peerId(peerId) {
      this.setQuery("peerId", peerId);
      this.reload();
    },
    channelId(channelId) {
      this.setQuery("channelId", channelId);
      this.reload();
    },
  },
  async mounted() {
    const [peersRes, channelsRes] = await Promise.all([
      this.$http.get("/peers"),
      this.$http.get("/channels"),
    ]);
    if (peersRes.data.peers) {
      this.peers.push(...peersRes.data.peers);
    }
    if (channelsRes.data.channels) {
      this.channels.push(...channelsRes.data.channels);
    }
    this.peerId = this.$route.query.peerId;
    this.channelId = this.$route.query.channelId;
    this.reload();
  },
  methods: {
    async reload() {
      const res = await this.$http.get(
        "/chaincodes" +
          this.encodeQuery({
            channelId: this.channelId,
            peerId: this.peerId,
          })
      );
      this.chaincodes.splice(0, this.chaincodes.length);
      if (res.data.chaincodes) {
        this.chaincodes.push(...res.data.chaincodes);
      }
    },
  },
};
</script>
