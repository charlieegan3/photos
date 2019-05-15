const Location = {
  template: `
    <span>{{ location }}</span>
`,
  data: function () {
    return {
      location: null
    }
  },
  created: function() {
    var app = this;
    axiosClient({
      method: "get",
      url: '/locations/' + app.$route.params.id + '.json'
    })
    .then(function(response) {
      app.location = response.data;
    })
    .catch(function(error) {
      console.log(error);
    })
  }
}
