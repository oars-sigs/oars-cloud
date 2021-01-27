<template>
  <v-app>
    <v-navigation-drawer v-model="drawer" clipped fixed app>
      <v-list dense>
        <v-list-group
          v-for="item in items"
          :key="item.title"
          v-model="item.active"
          :prepend-icon="item.action"
          no-action
        >
          <template v-slot:activator>
            <v-list-item-content>
              <v-list-item-title v-text="item.title"></v-list-item-title>
            </v-list-item-content>
          </template>

          <v-list-item
            v-for="subItem in item.items"
            :key="subItem.title"
            :to="subItem.path"
          >
            <v-list-item-content>
              <v-list-item-title v-text="subItem.title"></v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </v-list-group>
      </v-list>
    </v-navigation-drawer>
    <v-app-bar app fixed dark clipped-left color="indigo">
      <v-app-bar-nav-icon @click.stop="drawer = !drawer"></v-app-bar-nav-icon>
      <v-toolbar-title>OarsCloud</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-menu min-width="200" offset-y>
        <template v-slot:activator="{ on, attrs }">
          <v-avatar
            size="36"
            color="blue darken-1"
            dark
            v-bind="attrs"
            v-on="on"
          >
            <v-icon dark>mdi-account-circle</v-icon>
          </v-avatar>
        </template>
        <v-list dense>
          <v-list-item link>
            <v-list-item-content>
              <v-list-item-title>游客</v-list-item-title>
              <v-list-item-subtitle>anymous@example.com</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-divider></v-divider>
          </v-list-item>
          <v-list-item link>
            <v-list-item-icon>
              <v-icon>mdi-logout</v-icon>
            </v-list-item-icon>
            <v-list-item-title>退出</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </v-app-bar>
    <v-main>
      <v-container fluid>
        <router-view></router-view>
      </v-container>
    </v-main>
  </v-app>
</template>

<script>
export default {
  data: () => ({
    drawer: true,
    items: [
      {
        action: "mdi-view-dashboard",
        title: "集群",
        items: [
          { title: "集群概览", path: "/" },
          //{ title: "工作节点", path: "/node" },
          { title: "命名空间", path: "/namespace" },
        ],
      },
      {
        action: "mdi-web",
        title: "服务",
        items: [
          { title: "服务管理", path: "/service" },
          { title: "服务端点", path: "/endpoint" },
        ],
      },
      {
        action: "mdi-sitemap",
        title: "网关",
        items: [
          { title: "证书管理", path: "/cert" },
          { title: "端口管理", path: "/listener" },
          { title: "路由配置", path: "/route" },
        ],
      },
      {
        action: "mdi-settings",
        title: "配置",
        items: [
          { title: "配置管理", path: "/setting" },
        ],
      },
    ],
    itemIndex: 0,
  }),
  components: {},
  created(){
    let _that=this
    this.items.forEach((element,index )=> {
      element.items.forEach(element => {
          if (_that.$route.path==element.path){
            _that.items[index].active=true
          }
      });
    });
  },
};
</script>