<template>
  <v-container fluid class="px-0">
    <v-row>
      <v-col cols="12" sm="6" lg="3">
        <v-card class="mx-auto" max-width="344" outlined>
          <v-card-text>
            <div>CPU</div>
            <p class="text--primary text-center">
              <span class="display-1"> {{ cpuCore }} </span> cores
            </p>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" sm="6" lg="3">
        <v-card class="mx-auto" max-width="344" outlined>
          <v-card-text>
            <div>内存</div>
            <p class="text--primary text-center">
              <span class="display-1"> {{ memory }} </span> Gi
            </p>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" sm="6" lg="3">
        <v-card class="mx-auto" max-width="344" outlined>
          <v-card-text>
            <div>节点</div>
            <p class="text--primary text-center">
              <span class="display-1"> {{ nodeCount }} </span>
            </p>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" sm="6" lg="3">
        <v-card class="mx-auto" max-width="344" outlined>
          <v-card-text>
            <div>端点</div>
            <p class="text--primary text-center">
              <span class="display-1"> {{ edpCount }} </span> 
            </p>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
    <v-subheader>工作节点</v-subheader>
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
          <tr v-for="item in nodes" :key="item.status.id">
            <td>{{ item.status.id }}</td>
            <td>{{ item.status.ip }}</td>
            <td>{{ item.status.hostInfo.host.os }}</td>
            <td>{{ item.status.hostInfo.host.kernelArch }}</td>
            <td>{{ item.status.hostInfo.host.kernelVersion }}</td>
            <td>{{ item.status.hostInfo.logicalCores }}</td>
            <td>{{ (item.status.hostInfo.memory/1024/1024/1024).toFixed(2) }}Gi</td>
            <td :class="item.status.state+'-state'">{{ item.status.state }}</td>
          </tr>
        </tbody>
      </template>
    </v-simple-table>
  </v-container>
</template>

<script>
export default {
  components: {},
  data() {
    return {
      cpuCore: 0,
      memory: 0,
      nodeCount: 0,
      edpCount: 0,
      nodes: [],
    };
  },
  created() {
    this.list();
    this.listEdp();
  },
  methods: {
    list: function () {
      let _that = this;
       this.$call("system.admin.endpoint.get",{namespace:"system",service:"node"}).then((resp) => {
        _that.nodes = resp.data;
        _that.nodeCount=_that.nodes.length
        let memory =0, cpuCore =0
        _that.nodes.forEach(element => {
          memory+=element.status.hostInfo.memory;
          cpuCore+=element.status.hostInfo.logicalCores
        });
        _that.memory=(memory/1024/1024/1024).toFixed(2);
        _that.cpuCore=cpuCore;
      });
    },
    listEdp: function () {
      let _that = this;
      this.$call("system.admin.endpoint.get", {
      }).then((resp) => {
        _that.edpCount = resp.data.length;
      });
    },
  },
};
</script>

<style>
.v-progress-circular {
  margin: 1rem;
}
.running-state{
  color: green;
}
.error-state{
  color: red;
}
</style>