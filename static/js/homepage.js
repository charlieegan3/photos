const Homepage = {
  template: `
<div class="mw7 center cf">
  <div v-for="post in posts" class="fl w-third">
    <router-link class="db aspect-ratio aspect-ratio--1x1 dim" :to="{ name: 'posts.show', params: { id: post.id }}">
      <img v-lazy="post.url" class="bg-center cover aspect-ratio--object lazyload">
    </router-link>
  </div>
</div>
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
      app.posts = response.data.slice(0,10);
      for (var i = 0; i < app.posts.length; i++) {
        app.posts[i].url = "http://storage.googleapis.com/charlieegan3-instagram-archive/current/" + app.posts[i].id + ".jpg"
      }
    })
    .catch(function(error) {
      console.log(error);
    })
  }
}
