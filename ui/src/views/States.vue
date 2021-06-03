<style scoped>
.state-value {
  padding: 0 2em 2em 2em;
}
</style>

<template>
  <infinite-scroll @load-more="loadMore" :complete="statesComplete">
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
        <el-input
          v-model="transactionId"
          placeholder="Transaction ID"
          clearable
        ></el-input>
      </el-form-item>
      <el-form-item>
        <el-date-picker
          v-model="fromCreatedAtDate"
          type="datetime"
          placeholder="From created at"
          @change="reloadStates"
        >
        </el-date-picker>
      </el-form-item>
    </el-form>
    <el-table :data="states" stripe>
      <el-table-column prop="channelName" label="Channel"></el-table-column>
      <el-table-column
        prop="transactionId"
        label="Transaction ID"
      ></el-table-column>
      <el-table-column prop="key" label="Key"> </el-table-column>
      <el-table-column prop="type" label="Type"></el-table-column>
      <el-table-column
        prop="createdAtReadable"
        label="Created At"
      ></el-table-column>
      <el-table-column fixed="right" label="" width="120">
        <template #default="{ row }">
          <el-button @click="valueDrawer.state = row" type="text"
            >Details</el-button
          >
        </template>
      </el-table-column>
    </el-table>
  </infinite-scroll>
  <el-drawer
    :title="valueDrawer.title"
    v-model="valueDrawer.visible"
    :before-close="(valueDrawer.state = undefined)"
    direction="rtl"
    size="60%"
  >
    <div class="state-value">
      <json-viewer
        v-if="valueDrawer.value !== 'null'"
        :json="valueDrawer.value"
        root-name="value"
      />
      <div v-else>{{ valueDrawer.rawValue }}</div>
    </div>
  </el-drawer>
</template>

<script>
import JsonViewer from "@/components/JsonViewer";
export default {
  name: "States",
  components: { JsonViewer },
  data() {
    return {
      channels: [],
      channelsMap: {},
      channelId: this.$route.query.channelId,
      transactionId: this.$route.query.transactionId,
      fromCreatedAt: this.$route.query.fromCreatedAt,
      fromCreatedAtDate: this.parseDate(this.$route.query.fromCreatedAt),
      states: [],
      statesComplete: false,
      valueDrawer: {
        key: this.$route.query.showState,
        state: undefined,
        title: undefined,
        visible: false,
        value: undefined,
        rawValue: undefined,
      },
    };
  },
  watch: {
    fromCreatedAt(fromCreatedAt) {
      this.setQuery("fromCreatedAt", fromCreatedAt);
      this.fromCreatedAtDate = this.parseDate(fromCreatedAt);
    },
    channelId(channelId) {
      this.setQuery("channelId", channelId);
      this.reloadStates();
    },
    transactionId(transactionId) {
      this.setQuery("transactionId", transactionId.trim());
      this.reloadStates();
    },
    "valueDrawer.state"(state) {
      if (!state) {
        this.setQuery("showState", undefined);
        return;
      }
      this.setQuery("showState", state.key);
      this.valueDrawer.title = state.key;
      this.valueDrawer.visible = true;
      this.valueDrawer.value = atob(state.value);
      this.valueDrawer.rawValue = state.rawValue;
    },
  },
  async mounted() {
    const res = await this.$http.get("/channels");
    if (res.data.channels.length) {
      this.channels.push(...res.data.channels);
      this.channels.forEach((c) => (this.channelsMap[c.id] = c));
    }
    this.channelId = this.$route.query.channelId;
    await this.loadStates();
  },
  methods: {
    async loadStates(more) {
      let fromCreatedAt;
      if (more) {
        const lastState = this.states[this.states.length - 1];
        fromCreatedAt = lastState ? lastState.createdAt : undefined;
      } else {
        fromCreatedAt = this.fromCreatedAt;
      }
      const res = await this.$http.get(
        "/states" +
          this.encodeQuery({
            channelId: this.channelId,
            transactionId: (this.transactionId || "").trim(),
            fromCreatedAt: fromCreatedAt,
            loadMore: more,
          })
      );
      if (res.data.states.length) {
        this.states.push(
          ...res.data.states.map((s) => ({
            ...s,
            channelName: this.channelsMap[s.channelId].name,
            createdAtReadable: this.readableDate(this.parseDate(s.createdAt)),
          }))
        );
        this.fromCreatedAt = res.data.states[0].createdAt;
        this.statesComplete = false;
      } else {
        this.statesComplete = true;
      }
    },
    async loadMore() {
      await this.loadStates(true);
    },
    async reloadStates() {
      this.states.splice(0, this.states.length);
      this.fromCreatedAt = undefined;
      await this.loadStates();
    },
  },
};
</script>
