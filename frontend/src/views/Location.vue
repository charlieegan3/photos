<template>
  <div>
    <h1 class="dark-gray">photos taken in <span class="gray">{{ data.name }}</span></h1>
    <Grid v-if="items" :items="items"/>
  </div>
</template>

<script>
import Grid from '@/components/Grid.vue'
import axios from 'axios';

export default {
  name: 'home',
  created() {
    axios.get("//localhost:8000/locations/" + this.$route.params.id + ".json").then(({ data }) => {
      this.data = data;
    }).catch(function (error) {
      console.log(error);
    })
  },
  components: { Grid },
  watch: {
    data: function(data) {
		this.items = [];
		for (var i = 0; i < data.posts.length; i++) {
			this.items.push({
				post_id: data.posts[i].FullID,
				link: "/posts/" + data.posts[i].FullID,
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
