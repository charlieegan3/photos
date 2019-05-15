const Post = {
  template: `
    <div>
      <div v-if="post">
        <p>{{ post.caption }}</p>
        <router-link :to="{ name: 'locations.show', params: { id: post.location.id }}">
          <a>{{ post.location.id }}</a>
        </router-link>
      </div>
    </div>
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
