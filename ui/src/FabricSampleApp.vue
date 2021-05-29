<style>
html {
  font-family: "Helvetica Neue", Helvetica, "PingFang SC", "Hiragino Sans GB",
    "Microsoft YaHei", "微软雅黑", Arial, sans-serif;
}
</style>

<template>
  <el-container>
    <el-header>
      <el-menu
        mode="horizontal"
        class="menu"
        :router="true"
        :default-active="$route.name"
      >
        <el-menu-item index="assets">Assets</el-menu-item>
        <el-submenu index="system">
          <template #title>System</template>
          <el-menu-item index="peers">Peers</el-menu-item>
          <el-menu-item index="channels">Channels</el-menu-item>
          <el-menu-item index="chaincodes">Chaincodes</el-menu-item>
          <el-menu-item index="blocks">Blocks</el-menu-item>
          <el-menu-item index="states">States</el-menu-item>
        </el-submenu>
      </el-menu>
    </el-header>
    <el-main>
      <router-view />
    </el-main>
  </el-container>
</template>

<script>
export default {
  name: "FabricSampleApp",
};
</script>
