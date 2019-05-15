const Post = {
  template: `
    <span>{{ post.caption }}</span>
`,
  data: function () {
    return {
      post: null
    }
  },
  created: function() {
    var app = this;
    axiosClient({
      method: "get",
      url: '/posts/' + app.$route.params.id
    })
    .then(function(response) {
      app.post = response.data;
    })
    .catch(function(error) {
      console.log(error);
    })
  }
}
