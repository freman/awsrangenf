import Vue from 'vue'
import Router from 'vue-router'
import Dashboard from '@/components/Dashboard.vue'
import Config from '@/components/Config.vue'
import Imports from '@/components/Imports.vue'
import Custom from '@/components/Custom.vue'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      name: 'dashboard',
      component: Dashboard
    },
    {
      path: '/daemon',
      name: 'daemon',
      component: Config
    },
    {
      path: "/imports",
      name: 'imports',
      component: Imports,
    },
    {
      path: "/custom",
      name: 'custom',
      component: Custom,
    }
  ]
})
