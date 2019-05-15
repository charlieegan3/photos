var app = null;
function init() {
  app = new Vue(appConfig)
}

const axiosClient = axios.create({
  baseURL: '/', timeout: 1000
});

const routes = [
  { path: '/', component: Homepage },
  { name: 'posts.show', path: '/posts/:id', component: Post },
  { name: 'locations.show', path: '/locations/:id', component: Location }
]

var appConfig = {
  // router: new VueRouter({ routes: routes, mode: 'history' }),
  router: new VueRouter({ routes: routes }),
  el: '#app',
  data: {},
  methods: {},
  watch: {},
  computed: {},
  created() {}
}
