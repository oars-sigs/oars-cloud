<template>
  <v-container fluid class="px-0">
    <v-breadcrumbs :items="navs">
      <template v-slot:item="{ item }">
        <v-breadcrumbs-item :href="item.href" :disabled="item.disabled">
          {{ item.text.toUpperCase() }}
        </v-breadcrumbs-item>
      </template>
    </v-breadcrumbs>
    <v-simple-table class="float-none">
      <template v-slot:default>
        <thead>
          <tr>
            <th class="text-left">主机名</th>
            <th class="text-left">主机IP</th>
            <th class="text-left">系统</th>
            <th class="text-left">内核架构</th>
            <th class="text-left">内核版本</th>
            <th class="text-left">CPU</th>
            <th class="text-left">内存</th>
            <th >状态</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in nodes" :key="item.hostname">
            <td>{{ item.hostname }}</td>
            <td>{{ item.hostIP }}</td>
            <td>{{ item.hostInfo.host.os }}</td>
            <td>{{ item.hostInfo.host.kernelArch }}</td>
            <td>{{ item.hostInfo.host.kernelVersion }}</td>
            <td>{{ item.hostInfo.logicalCores }}</td>
            <td>{{ (item.hostInfo.memory/1024/1024/1024).toFixed(2) }}Gi</td>
            <td>{{ item.status }}</td>
          </tr>
        </tbody>
      </template>
    </v-simple-table>
    <v-overlay :value="overlay">
      <v-progress-circular indeterminate size="64"></v-progress-circular>
    </v-overlay>
  </v-container>
</template>

<script>
import { fetch } from "@/utils/request";
export default {
  data() {
    return {
      dialog: false,
      delDialog: false,
      overlay: false,
      navs:[
           {
              text: "主页",
              href: "/"
          },
          {
              text: "节点"
          }
      ],
      nodes: [],
      actions: [
      ],
      actionParam: {},
    };
  },
  created() {
    this.initActionParam();
    this.overlay = true;
    this.list();
  },
  methods: {
    list: function () {
      let _that = this;
      fetch("system.admin.endpoint.get",{namespace:"system",service:"node"}).then((resp) => {
        _that.nodes = resp.data;
        this.overlay = false;
      });
    },
    initActionParam: function () {
      this.actionParam = {
        args: {},
      };
    },
    doAction: function () {
    },
  },
};
</script>

<style>
</style>