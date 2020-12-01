<template>
  <v-container fluid class="px-0">
    <v-breadcrumbs :items="navs" class="px-0">
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
          v-model="actionParam.args.namespace"
          @change="
            actionParam.key = 'list';
            doAction();
          "
          label="命名空间"
          dense
        ></v-select>
      </v-col>
      <v-col cols="2" align="right">
        <v-dialog v-model="dialog" persistent max-width="600px">
          <template v-slot:activator="{ on, attrs }">
            <v-btn color="primary" dark v-bind="attrs" @click="actionTitle='创建';update=false;svcYaml='';selectNodes=[]" text v-on="on">
              <v-icon left>mdi-plus</v-icon> 创建
            </v-btn>
          </template>
          <v-card>
            <v-card-title>
              <span>{{ actionTitle }}服务</span>
            </v-card-title>
            <v-card-text>
              <v-container>
                <v-row>
                  <v-col cols="12">
                    <v-text-field
                      label="Name*"
                      v-model="actionParam.args.name"
                      hint="仅支持小写字母、数字和‘-’"
                      required
                      :disabled="update"
                    ></v-text-field>
                  </v-col>
                  <!-- <v-col cols="12">
                    <v-select
                      :items="kinds"
                      v-model="actionParam.args.kind"
                      @change="
                        actionParam.key = 'list';
                        doAction();
                      "
                      label="类型"
                      dense
                      :disabled="update"
                    ></v-select>
                  </v-col> -->
                  <v-col cols="12">
                    <v-select
                      :items="nodes"
                      v-model="selectNodes"
                      label="端点"
                      dense
                      multiple
                    ></v-select>
                  </v-col>
                  <v-col cols="12">
                    <Editor
                      class="editor"
                      v-model="svcYaml"
                      @init="editorInit"
                      lang="yaml"
                      theme="chrome"
                      height="300"
                    ></Editor>
                  </v-col>
                </v-row>
              </v-container>
            </v-card-text>
            <v-card-actions>
              <v-spacer></v-spacer>
              <v-btn
                color="blue darken-1"
                text
                @click="
                  dialog = false;
                  initActionParam();
                "
                >取消</v-btn
              >
              <v-btn
                color="blue darken-1"
                text
                @click="
                  actionParam.key = 'create';
                  doAction();
                "
                >{{ actionTitle }}</v-btn
              >
            </v-card-actions>
          </v-card>
        </v-dialog>
      </v-col>
    </v-row>
    <v-simple-table>
      <template v-slot:default>
        <thead>
          <tr>
            <th class="text-left">名称</th>
            <th class="text-right">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in services" :key="item.name">
            <td>{{ item.name }}</td>
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
        <v-card-text
          >确定删除服务‘{{ actionParam.args.name }}’?</v-card-text
        >
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
import Editor from "vue2-ace-editor";
import yaml from "js-yaml";
import YAML  from "json2yaml";
export default {
  components: {
    Editor,
  },
  data() {
    return {
      dialog: false,
      delDialog: false,
      overlay: false,
      actionTitle:"管理",
      update: false,
      navs: [
        {
          text: "主页",
          href: "/",
        },
        {
          text: "服务管理",
        },
      ],
      services: [],
      namespaces: [],
      namespace: "",
      actions: [
        {
          title: "编辑",
          icon: "mdi-square-edit-outline",
          key: "edit",
        },
        {
          title: "删除",
          icon: "mdi-trash-can-outline",
          color: "red",
          key: "confirmDelete",
        },
      ],
      actionParam: {
        args: {},
      },
      svcYaml: "",
      kinds:["docker"],
      nodes: [],
      selectNodes:[],
    };
  },
  created() {
    this.initActionParam();
    this.overlay = true;
    this.listNamespace();
    this.listNode();
  },
  methods: {
    initActionParam: function () {
      let namespace = this.actionParam.args.namespace;
      this.actionParam = {
        args: {
          namespace: namespace,
        },
      };
    },
    editorInit: function () {
      require("brace/ext/language_tools");
      require("brace/theme/chrome");
      require("brace/mode/yaml");
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
        _that.actionParam.args.namespace = _that.$store.state.currentNamespace;
        _that.list();
      });
    },
    listNode: function () {
      let _that = this;
      this.$call("system.admin.endpoint.get",{namespace:"system",service:"node"}).then((resp) => {
        resp.data.forEach((element) => {
          _that.nodes.push(element.hostname);
        });
        this.overlay = false;
      });
    },
    list: function () {
      let _that = this;
      this.$call("system.admin.service.get", {
        namespace: this.actionParam.args.namespace,
      }).then((resp) => {
        _that.services = resp.data;
        this.overlay = false;
      });
    },
    doAction: function () {
      let _that = this;
     
      switch (this.actionParam.key) {
        case "list":
          _that.$store.commit('SetCurrentNamespace',_that.actionParam.args.namespace);
          this.overlay = true;
          _that.list();
          break;
        case "confirmDelete":
          _that.delDialog = true;
          break;
        case "delete":
          this.overlay = true;
          _that.delDialog = false;
          this.$call("system.admin.service.delete", this.actionParam.args).then(
            () => {
              _that.list();
            }
          );
          _that.initActionParam();
          break;
        case "create":
          this.overlay = true;
          var content=yaml.safeLoad(_that.svcYaml)
          this.actionParam.args.kind="docker"
          this.actionParam.args[_that.actionParam.args.kind]= content;
          _that.actionParam.args.endpoints=new Array();
          _that.selectNodes.forEach(function(item){
            if (item)
            _that.actionParam.args.endpoints.push({hostname: item});
          })
          this.$call("system.admin.service.put", this.actionParam.args).then(
            () => {
              _that.initActionParam();
              _that.dialog = false;
              _that.list();
            }
          );
          break;
        case "edit":
          _that.actionTitle="更新";
          _that.dialog=true;
          _that.update=true;
          _that.selectNodes=new Array();
          _that.actionParam.args.endpoints.forEach(function(item){
            _that.selectNodes.push(item.hostname)
          })
         _that.svcYaml=YAML.stringify(_that.actionParam.args[_that.actionParam.args.kind])
      }
    },
  },
};
</script>

<style>
</style>