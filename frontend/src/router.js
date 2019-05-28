import Vue from 'vue'
import Router from 'vue-router'
import Home from './views/Home.vue'

Vue.use(Router)

export default new Router({
  mode: 'history',
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home
    },
    {
      path: '/tags',
      name: 'tags',
      component: () => import(/* webpackChunkName: "tags" */ './views/Tags.vue')
    },
    {
      path: '/locations',
      name: 'locations',
      component: () => import(/* webpackChunkName: "locations" */ './views/Locations.vue')
    },
    {
      path: '/calendar',
      name: 'calendar',
      component: () => import(/* webpackChunkName: "calendar" */ './views/Calendar.vue')
    },
    {
      path: '/posts/:id',
      name: 'post',
      component: () => import(/* webpackChunkName: "posts.show" */ './views/Post.vue')
    },
    {
      path: '/tags/:id',
      name: 'tag',
      component: () => import(/* webpackChunkName: "posts.show" */ './views/Tag.vue')
    },
    {
      path: '/locations/:id',
      name: 'location',
      component: () => import(/* webpackChunkName: "posts.show" */ './views/Location.vue')
    },
  ]
})
