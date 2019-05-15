const Homepage = {
  template: `
<ul>
  <li v-for="post in posts">
    <router-link :to="{ name: 'posts.show', params: { id: post.id }}">
      <a>{{ post.id }}</a>
    </router-link>
  </li>
</ul>
`,
  data: function () {
    return {
      posts: []
    }
  },
  created: function() {
    var app = this;
    axiosClient({
      method: "get",
      url: '/index.json',
    })
    .then(function(response) {
      app.posts = response.data;
    })
    .catch(function(error) {
      console.log(error);
    })
  }
}
