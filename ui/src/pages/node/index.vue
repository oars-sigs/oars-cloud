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
          <tr v-for="item in nodes" :key="item.status.id">
            <td>{{ item.status.id }}</td>
            <td>{{ item.status.ip }}</td>
            <td>{{ item.status.hostInfo.host.os }}</td>
            <td>{{ item.status.hostInfo.host.kernelArch }}</td>
            <td>{{ item.status.hostInfo.host.kernelVersion }}</td>
            <td>{{ item.status.hostInfo.logicalCores }}</td>
            <td>{{ (item.status.hostInfo.memory/1024/1024/1024).toFixed(2) }}Gi</td>
            <td :class="item.status.state+'-state'">{{ item.status.state }}</td>
            <td class="text-right">
              <v-menu bottom :offset-x="true" :offset-y="true">
                <template v-slot:activator="{ on, attrs }">
                  <v-btn icon v-bind="attrs" v-on="on">
                    <v-icon>mdi-dots-vertical</v-icon>
                  </v-btn>
                </template>
                <v-list dense>
                  <v-list-item
                    v-for="(action, i) in actions"
                    :key="i"
                    @click="
                      actionParam.key = action.key;
                      actionParam.args = item;
                      doAction();
                    "
                  >
                    <v-list-item-title>
                      <v-icon dense :color="action.color">{{
                        action.icon
                      }}</v-icon>
                      {{ action.title }}
                    </v-list-item-title>
                  </v-list-item>
                </v-list>
              </v-menu>
            </td>
          </tr>
        </tbody>
      </template>
    </v-simple-table>
    <v-dialog v-model="delDialog" persistent max-width="290">
      <v-card>
        <v-card-title></v-card-title>
        <v-card-text>确定删除‘{{ actionParam.args.name }}’?</v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            text
            @click="
              delDialog = false;
              initActionParam();
            "
            >取消</v-btn
          >
          <v-btn
            color="red"
            text
            @click="
              actionParam.key = 'delete';
              doAction();
            "
            >确定</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>
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
        {
          title: "删除",
          icon: "mdi-trash-can-outline",
          color: "red",
          key: "confirmDelete",
        },
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
      let _that = this;
      switch (this.actionParam.key) {
        case "confirmDelete": {
          _that.delDialog = true;
          break;
        }
        case "delete": {
          this.overlay = true;
          _that.delDialog = false;
          this.$call(
            "system.admin.endpoint.delete",
            this.actionParam.args
          ).then(() => {
            _that.list();
          });
          _that.initActionParam();
          break;
        }
      }
    },
  },
};
</script>

<style>
.running-state{
  color: green;
}
.error-state{
  color: red;
}
</style>