import Vue from 'vue'
import Router from 'vue-router'
import Index from '@/pages/index/index'
import Main from '@/pages/index/main'
import Setting from '@/pages/setting/index'
import Service from '@/pages/service/index'
import Endpoint from '@/pages/service/endpoint' 
import EndpointExec from '@/pages/service/exec' 
import Namespace from '@/pages/namespace/index'
import Node from '@/pages/node/index'
import Listener from '@/pages/ingress/listener'
import Route from '@/pages/ingress/route'
import Cert from '@/pages/ingress/cert'
import Monitor from '@/pages/monitor/index'

Vue.use(Router)

let router = new Router({
  mode: "",
  base: "/",
  routes: [
    {
      path: '/',
      name: 'Main',
      component: Main,
      children: [
        {
          path: '/',
          name: 'Index',
          component: Index, 
          meta:{
            title: '集群概览'
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
        {
          path: '/exec',
          name: 'EndpointExec',
          component: EndpointExec,
          meta:{
            title: '端点'
          }
        },
        {
          path: '/cert',
          name: 'Cert',
          component: Cert,
          meta:{
            title: '证书管理'
          }
        },
        {
          path: '/listener',
          name: 'Listener',
          component: Listener,
          meta:{
            title: '端口管理'
          }
        },
        {
          path: '/route',
          name: 'Route',
          component: Route,
          meta:{
            title: '路由配置'
          }
        },
        {
          path: '/monitor',
          name: 'Monitor',
          component: Monitor,
          meta:{
            title: '监控管理'
          }
        },
      ]
    },
    
    ]
})


export default router

