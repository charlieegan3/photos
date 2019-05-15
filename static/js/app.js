var app = null;
function init() {
  app = new Vue(appConfig)
}

const axiosClient = axios.create({
  baseURL: '/', timeout: 1000
});

const routes = [
  { path: '/', component: Homepage },
  { name: 'posts.show', path: '/posts/:id', component: Post }
]

var appConfig = {
  // router: new VueRouter({ routes: routes, mode: 'history' }),
  router: new VueRouter({ routes: routes }),
  el: '#app',
  data: {
    things: "charlie",
  },
  methods: {
    lastDate: function() {
    },
  },
  watch: {
    date: function() {
      this.updateEntries()
    }
  },
  computed: {
    groups: function() {
    }
  },
  created() {
  }
}
