<style scoped>
.channel-config {
  padding: 0 2em 2em 2em;
}
</style>

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
  <el-table :data="channels" ref="channelsTable">
    <el-table-column prop="id" label="ID"> </el-table-column>
    <el-table-column prop="name" label="Name"> </el-table-column>
    <el-table-column fixed="right" label="" width="120">
      <template #default="{ row }">
        <el-button @click="showConfig(row)" type="text">Configs</el-button>
      </template>
    </el-table-column>
  </el-table>
  <el-drawer
    :title="configsDrawer.title"
    v-model="configsDrawer.visible"
    :before-close="(configsDrawer.channel = undefined)"
    direction="rtl"
    size="60%"
  >
    <div class="channel-config">
      <json-viewer :js="configsDrawer.configs" root-name="value" />
    </div>
  </el-drawer>
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
      configsDrawer: {
        channel: undefined,
        title: undefined,
        visible: false,
        configs: undefined,
      },
    };
  },
  watch: {
    peerId(peerId) {
      this.setQuery("peerId", peerId);
      this.reload();
    },
    "configsDrawer.channel"(channel) {
      if (!channel) {
        this.setQuery("showConfigs", undefined);
        return;
      }
      this.setQuery("showConfigs", channel.name);
      this.configsDrawer.title = channel.name;
      this.configsDrawer.visible = true;
      this.configsDrawer.configs = channel.configs;
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
    async showConfig(channel) {
      if (!channel.configs) {
        const res = await this.$http.get(
          "/channel_configs?channelId=" + channel.id
        );
        if (res.data.channelConfigs.length) {
          channel.configs = res.data.channelConfigs.map((cc) =>
            JSON.parse(atob(cc.parsed))
          );
        }
      }
      this.configsDrawer.channel = channel;
    },
  },
};
</script>
