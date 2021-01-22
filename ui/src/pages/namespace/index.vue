<template>
  <v-container fluid class="px-0">
    <v-breadcrumbs :items="navs">
      <template v-slot:item="{ item }">
        <v-breadcrumbs-item :href="item.href" :disabled="item.disabled">
          {{ item.text.toUpperCase() }}
        </v-breadcrumbs-item>
      </template>
    </v-breadcrumbs>
    <v-col align="right">
      <v-dialog v-model="dialog" persistent max-width="600px">
        <template v-slot:activator="{ on, attrs }">
          <v-btn
            color="primary"
            align="right"
            dark
            v-bind="attrs"
            text
            v-on="on"
          >
            <v-icon left>mdi-plus</v-icon> 创建
          </v-btn>
        </template>
        <v-card>
          <v-card-title>
            <span>添加命名空间</span>
          </v-card-title>
          <v-card-text>
            <v-container>
              <v-row>
                <v-col cols="12">
                  <v-text-field
                    label="Name*"
                    v-model="actionParam.args.name"
                    hint="仅支持字母、数字和‘-’"
                    required
                  ></v-text-field>
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
              >创建</v-btn
            >
          </v-card-actions>
        </v-card>
      </v-dialog>
    </v-col>
    <v-simple-table class="float-none">
      <template v-slot:default>
        <thead>
          <tr>
            <th class="text-left">名称</th>
            <th class="text-right">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in namespaces" :key="item.name">
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
                      actionParam.args.name = item.name;
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
          >确定删除命名空间‘{{ actionParam.args.name }}’?</v-card-text
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
    <v-overlay :absolute="true" :value="overlay">
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
              text: "命名空间"
          }
      ],
      namespaces: [],
      actions: [
        {
          title: "编辑",
          icon: "mdi-square-edit-outline",
        },
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
      fetch("system.admin.namespace.get").then((resp) => {
        _that.namespaces = resp.data;
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
        case "confirmDelete":
          _that.delDialog = true;
          break;
        case "delete":
          this.overlay = true;
          _that.delDialog = false;
          fetch("system.admin.namespace.delete", this.actionParam.args).then(() => {
            _that.list();
          });
          _that.initActionParam();
          break;
        case "create":
          this.overlay = true;
          fetch("system.admin.namespace.put", this.actionParam.args).then(() => {
            _that.initActionParam();
            _that.dialog = false;
            _that.list();
          });
      }
    },
  },
};
</script>

<style>
</style>