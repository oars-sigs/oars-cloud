<template>
  <v-container fluid class="px-0">
    <v-breadcrumbs :items="navs">
      <template v-slot:item="{ item }">
        <v-breadcrumbs-item :href="item.href" :disabled="item.disabled">
          {{ item.text.toUpperCase() }}
        </v-breadcrumbs-item>
      </template>
    </v-breadcrumbs>
    <v-row>
      <v-col cols="8"></v-col>
      <v-col cols="2">
        <v-select
          :items="namespaces"
          v-model="namespace"
          label="命名空间"
          @change="listService"
          dense
        ></v-select>
      </v-col>
      <v-col cols="2">
        <v-select
          :items="services"
          v-model="service"
          @change="list"
          label="服务"
          dense
        ></v-select>
      </v-col>
    </v-row>

    <v-simple-table class="float-none">
      <template v-slot:default>
        <thead>
          <tr>
            <th class="text-left">端点</th>
            <th class="text-left">命名空间</th>
            <th class="text-left">服务</th>
            <th class="text-left">节点</th>
            <th>状态</th>
            <th>创建时间</th>
            <th class="text-right">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in endpoints" :key="item.id">
            <td>{{ item.name }}</td>
            <td>{{ item.namespace }}</td>
            <td>{{ item.service }}</td>
            <td>{{ item.hostname }}</td>
            <td>
              <v-tooltip right>
                <template v-slot:activator="{ on, attrs }">
                  <span v-on="on" v-bind="attrs">{{ item.state }}</span>
                </template>
                <span>{{ item.status }}</span>
              </v-tooltip>
            </td>
            <td>{{ item.created | formatT }}</td>
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
    <v-overlay :value="overlay">
      <v-progress-circular indeterminate size="64"></v-progress-circular>
    </v-overlay>
    <v-dialog
      v-model="logDialog"
      persistent
      :scrollable="true"
      max-width="1000"
    >
      <v-card>
        <v-card-title>服务‘{{ actionParam.args.name }}’日志</v-card-title>
        <v-card-text
          style="white-space: pre-wrap; background-color: black; color: #fff"
        >
          {{ logs }}
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            text
            @click="
              logDialog = false;
              initActionParam();
              term.dispose();
            "
            >取消</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script>
import { formatDate } from "../../utils/formatDate.js";
export default {
  data() {
    return {
      dialog: false,
      logDialog: false,
      overlay: false,
      navs: [
        {
          text: "主页",
          href: "/",
        },
        {
          text: "服务端点",
        },
      ],
      actions: [
        {
          title: "重启",
          icon: "mdi-restart",
          color: "green",
          key: "restart",
        },
        {
          title: "停止",
          icon: "mdi-stop",
          color: "red",
          key: "stop",
        },
        {
          title: "日志",
          icon: "mdi-file-document",
          key: "log",
        },
        {
          title: "终端",
          icon: "mdi-powershell",
          key: "shell",
        },
      ],
      namespace: "",
      service: "",
      namespaces: [],
      services: [],
      endpoints: [],
      actionParam: {},
      logs: "",
      term: {},
    };
  },
  filters: {
    formatT(time) {
      if (time){
        time = time * 1000;
        let date = new Date(time);
        return formatDate(date, "yyyy-MM-dd hh:mm");
      }
      return "--";
    },
  },
  created() {
    this.initActionParam();
    this.overlay = true;
    this.listNamespace();
    this.list();
  },
  methods: {
    list: function () {
      let _that = this;
      this.$call("system.admin.endpoint.get", {
        namespace: this.namespace,
        service: this.service,
      }).then((resp) => {
        _that.endpoints = resp.data;
        this.overlay = false;
      });
    },
    listNamespace: function () {
      let _that = this;
      this.$call("system.admin.namespace.get").then((resp) => {
        let ns = new Array();
        resp.data.forEach((element) => {
          ns.push(element.name);
        });
        _that.namespaces = ns;
        if (!_that.$store.state.currentNamespace){
           _that.$store.commit('SetCurrentNamespace',ns[0]);
        }
        _that.namespace = _that.$store.state.currentNamespace;
        _that.listService();
      });
    },
    listService: function () {
      let _that = this;
      _that.$store.commit('SetCurrentNamespace',this.namespace);
      this.$call("system.admin.service.get", {
        namespace: this.namespace,
      }).then((resp) => {
        _that.services = [{ text: "All", value: "" }];
        resp.data.forEach((element) => {
          _that.services.push({ text: element.name, value: element.name });
        });
        _that.service = _that.services[0].value;
        _that.list();
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
        case "restart":
          _that.overlay = true;
          this.$call(
            "system.admin.endpoint.restart",
            this.actionParam.args
          ).then(() => {
            _that.initActionParam();
            _that.list();
          });
          break;
        case "stop":
          _that.overlay = true;
          this.$call("system.admin.endpoint.stop", this.actionParam.args).then(
            () => {
              _that.initActionParam();
              _that.list();
            }
          );
          break;
        case "log":
          var param = {
            id: this.actionParam.args.id,
            hostname:  this.actionParam.args.hostname,
            tail: "100",
          };
          this.$call("system.admin.endpoint.log", param).then((resp) => {
            _that.logDialog = true;
            var data = resp.data.replace(/[\r]/g, "\n");
            data = data.substring(8);
            _that.logs = data.replace(/\n(.{8})/g, "\r\n");
            // _that.logs="sss<br/>sss"
          });
          break;
        case "shell":{
          let queryParam = {
            hostname:  this.actionParam.args.hostname,     
            id: this.actionParam.args.id, 
            name: this.actionParam.args.name, 
          }
          this.$router.push({ 
            path:'/exec',  
            query:queryParam,
          });
          break;
        }
      }
    },
  },
};
</script>

<style>
</style>