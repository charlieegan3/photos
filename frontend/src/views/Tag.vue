<template>
  <div>
    <h1 class="dark-gray">photos tagged <span class="gray">#{{ data.name }}</span></h1>
    <Map v-if="items" :items="items" :height="400"/>
    <Grid v-if="items" :items="items"/>
  </div>
</template>

<script>
import axios from 'axios';
import Grid from '@/components/Grid.vue'
import Map from '@/components/Map.vue'

export default {
  name: 'home',
  created() {
    axios.get("/data/tags/" + this.$route.params.id + ".json").then(({ data }) => {
      this.data = data;
    }).catch(function (error) {
      console.log(error);
    })
  },
  components: { Grid, Map },
  watch: {
    data: function(data) {
      this.items = [];
      for (var i = 0; i < data.posts.length; i++) {
        var post = data.posts[i];
        this.items.push({
          post_id: post.FullID,
          link: "/posts/" + post.FullID,
          location: {
            name: post.location.name,
            lat: post.lat,
            long: post.long,
          }
        })
      }
    }
  },
  data() {
    return {
      data: false,
      items: false,
    }
  }
}
</script>

<style>
.post-container {
  width: 300px;
  height: 300px;
}
</style>
