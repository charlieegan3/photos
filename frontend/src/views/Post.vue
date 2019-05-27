<template>
  <div>
    <div class="post-container">
      <VideoItem v-if="data.is_video" :post="this.$route.params.id"/>
      <PhotoItem v-else :post="this.$route.params.id"/>
    </div>
    <p>
      {{ data.caption }}
    </p>
    <p>
      {{ data.tags }}
    </p>
    <p>
      {{ data.location }}
    </p>
  </div>
</template>

<script>
import axios from 'axios';
import PhotoItem from '@/components/PhotoItem.vue'
import VideoItem from '@/components/VideoItem.vue'

export default {
  name: 'home',
  created() {
    axios.get("//localhost:8000/posts/" + this.$route.params.id + ".json").then(({ data }) => {
      this.data = data;
    }).catch(function (error) {
      console.log(error);
    })
  },
  components: {
    VideoItem, PhotoItem,
  },
  data() {
    return {
      data: false
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
