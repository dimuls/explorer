<template>
  <infinite-scroll @load-more="loadMore" :complete="transactionsComplete">
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
      <el-form-item>
        <el-input-number
          v-model="blockId"
          placeholder="Block ID"
          :min="1"
          controls-position="right"
        ></el-input-number>
      </el-form-item>
      <el-form-item>
        <el-date-picker
          v-model="fromCreatedAtDate"
          type="datetime"
          placeholder="From created at"
          @change="reloadTransactions"
        >
        </el-date-picker>
      </el-form-item>
    </el-form>
    <el-table :data="transactions" stripe>
      <el-table-column prop="channelName" label="Channel"></el-table-column>
      <el-table-column prop="blockId" label="Block ID"></el-table-column>
      <el-table-column prop="id" label="ID"> </el-table-column>
      <el-table-column
        prop="createdAtReadable"
        label="Created At"
      ></el-table-column>
    </el-table>
  </infinite-scroll>
</template>

<script>
export default {
  name: "Transactions",
  data() {
    let blockId;
    if (this.$route.query.blockId) {
      blockId = parseInt(this.$route.query.blockId);
    }
    return {
      channels: [],
      channelsMap: {},
      channelId: undefined,
      blockId: blockId,
      fromCreatedAt: this.$route.query.fromCreatedAt,
      fromCreatedAtDate: this.parseDate(this.$route.query.fromCreatedAt),
      transactionsComplete: false,
      transactions: [],
    };
  },
  watch: {
    fromCreatedAt(fromCreatedAt) {
      this.setQuery("fromCreatedAt", fromCreatedAt);
      this.fromCreatedAtDate = this.parseDate(fromCreatedAt);
    },
    channelId(channelId) {
      this.setQuery("channelId", channelId);
      this.reloadTransactions();
    },
    blockId(blockId) {
      this.setQuery("blockId", blockId);
      this.reloadTransactions();
    },
  },
  async mounted() {
    const res = await this.$http.get("/channels");
    if (res.data.channels.length) {
      this.channels.push(...res.data.channels);
      this.channels.forEach((c) => (this.channelsMap[c.id] = c));
    }
    this.channelId = this.$route.query.channelId;
    await this.loadTransactions();
  },
  methods: {
    async loadTransactions(more) {
      let fromCreatedAt;
      if (more) {
        const lastTransaction = this.transactions[this.transactions.length - 1];
        fromCreatedAt = lastTransaction ? lastTransaction.createdAt : undefined;
      } else {
        fromCreatedAt = this.fromCreatedAt;
      }
      const res = await this.$http.get(
        "/transactions" +
          this.encodeQuery({
            channelId: this.channelId,
            blockId: this.blockId,
            fromCreatedAt: fromCreatedAt,
            loadMore: more,
          })
      );
      if (res.data.transactions.length) {
        this.transactions.push(
          ...res.data.transactions.map((t) => ({
            ...t,
            channelName: this.channelsMap[t.channelId].name,
            createdAtReadable: this.readableDate(this.parseDate(t.createdAt)),
          }))
        );
        this.fromCreatedAt = res.data.transactions[0].createdAt;
        this.transactionsComplete = false;
      } else {
        this.transactionsComplete = true;
      }
    },
    async loadMore() {
      await this.loadTransactions(true);
    },
    async reloadTransactions() {
      this.transactions.splice(0, this.transactions.length);
      this.fromCreatedAt = undefined;
      await this.loadTransactions();
    },
  },
};
</script>
