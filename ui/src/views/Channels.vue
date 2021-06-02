<template>
  <el-form :inline="true">
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
  </el-form>
  <el-table
    :data="channels"
    @row-click="toggleExpandChannelRow"
    @expand-change="channelRowExpandToggled"
    ref="channelsTable"
  >
    <el-table-column type="expand">
      <template #default="{ row }">
        <pre>
          {{ channelsConfigs[row.id] }}
        </pre>
      </template>
    </el-table-column>
    <el-table-column prop="id" label="ID"> </el-table-column>
    <el-table-column prop="name" label="Name"> </el-table-column>
  </el-table>
</template>

<script>
export default {
  name: "Channels",
  data() {
    return {
      peers: [],
      peerId: undefined,
      channels: [],
      channelsConfigs: {},
    };
  },
  watch: {
    peerId(peerId) {
      this.setQuery("peerId", peerId);
      this.reload();
    },
  },
  async mounted() {
    const res = await this.$http.get("/peers");
    if (res.data.peers) {
      this.peers.push(...res.data.peers);
    }
    this.peerId = this.$route.query.peerId;
    this.reload();
  },
  methods: {
    async reload() {
      const res = await this.$http.get(
        "/channels" +
          this.encodeQuery({
            peerId: this.peerId,
          })
      );
      this.channels.splice(0, this.channels.length);
      if (res.data.channels.length) {
        this.channels.push(...res.data.channels);
      }
    },
    async getChannelConfigs(channelId) {
      let configs = this.channelsConfigs[channelId];
      if (configs) {
        return configs;
      }
      const res = await this.$http.get(
        "/channel_configs?channelId=" + channelId
      );
      if (res.data.channelConfigs.length) {
        configs = res.data.channelConfigs;
        this.channelsConfigs[channelId] = configs;
      }
      return configs;
    },
  },
};
</script>
