import Vue from 'vue'
import Router from 'vue-router'
import Index from '@/pages/index/index'
import Setting from '@/pages/setting/index'
import Service from '@/pages/service/index'
import Endpoint from '@/pages/service/endpoint' 
import Namespace from '@/pages/namespace/index'
import Node from '@/pages/node/index'

Vue.use(Router)

let router = new Router({
  mode: '',
  routes: [
    {
      path: '/',
      name: 'Index',
      component: Index,
      meta:{
        title: '主页'
      }
    },
    {
      path: '/service',
      name: 'Service',
      component: Service,
      meta:{
        title: '服务'
      }
    },
    {
      path: '/setting',
      name: 'Setting',
      component: Setting,
      meta:{
        title: '设置'
      }
    },
    {
      path: '/namespace',
      name: 'Namespace',
      component: Namespace,
      meta:{
        title: '命名空间'
      }
    },
    {
      path: '/node',
      name: 'Node',
      component: Node,
      meta:{
        title: '节点'
      }
    },
    {
      path: '/endpoint',
      name: 'Endpoint',
      component: Endpoint,
      meta:{
        title: '端点'
      }
    },
    ]
})


export default router

