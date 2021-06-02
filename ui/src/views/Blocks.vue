<template>
  <infinite-scroll @load-more="loadBlocks" :complete="lastBlockId === 1">
    <el-form :inline="true">
      <el-form-item>
        <el-input-number
          v-model="fromId"
          placeholder="Block ID"
          :min="1"
          controls-position="right"
        ></el-input-number>
      </el-form-item>
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
    <el-table :data="blocks" stripe>
      <el-table-column prop="id" label="ID"> </el-table-column>
      <el-table-column prop="channelName" label="Channel"></el-table-column>
      <el-table-column prop="number" label="Number"></el-table-column>
    </el-table>
  </infinite-scroll>
</template>

<script>
export default {
  name: "Blocks",
  components: {},
  data() {
    const fromId = parseInt(this.$route.query.fromId, 10) || undefined;
    return {
      lastBlockId: fromId,
      fromId: fromId,
      channels: [],
      channelsMap: {},
      channelId: this.$route.query.channelId,
      blocks: [],
      noResult: false,
    };
  },
  watch: {
    lastBlockId(_, lastBlockIdBefore) {
      this.setQuery("fromId", lastBlockIdBefore);
    },
    fromId(fromId) {
      this.setQuery("fromId", fromId);
      this.reloadBlocks();
    },
    channelId(channelId) {
      this.setQuery("channelId", channelId);
      this.reload();
    },
  },
  async mounted() {
    const res = await this.$http.get("/channels");
    if (res.data.channels.length) {
      this.channels.push(...res.data.channels);
      this.channels.forEach((c) => (this.channelsMap[c.id] = c));
    }
    this.channelId = this.$route.query.channelId;
    await this.loadBlocks();
  },
  methods: {
    async loadBlocks() {
      const res = await this.$http.get(
        "/blocks" +
          this.encodeQuery({
            fromId: this.lastBlockId,
            channelId: this.channelId,
          })
      );
      if (res.data.blocks.length) {
        this.lastBlockId = parseInt(
          res.data.blocks[res.data.blocks.length - 1].id,
          10
        );
        this.blocks.push(
          ...res.data.blocks.map((b) => ({
            ...b,
            channelName: this.channelsMap[b.channelId].name,
          }))
        );
      }
    },
    async reloadBlocks() {
      this.blocks.splice(0, this.blocks.length);
      await this.loadBlocks();
    },
  },
};
</script>
