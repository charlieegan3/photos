var app = null;
function init() {
  app = new Vue(appConfig)
}

const axiosClient = axios.create({
  baseURL: '/', timeout: 1000
});

const Foo = { template: '<div>foo</div>' }
const Bar = { template: '<div>bar</div>' }

const routes = [
  { path: '/', component: Foo },
  { path: '/bar', component: Bar }
]


var appConfig = {
	router: new VueRouter({ routes: routes, mode: 'history' }),
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
