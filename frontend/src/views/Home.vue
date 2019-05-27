<template>
  <div class="home">
    <PostList v-if="posts" :posts="posts"/>
  </div>
</template>

<script>
import PostList from '@/components/PostList.vue'
import axios from 'axios';

export default {
  name: 'home',
  created() {
    axios.get("//localhost:8000/index.json").then(({ data }) => {
      this.posts = data;
    }).catch(function (error) {
      console.log(error);
    })
  },
  data() {
    return {
      posts: false
    }
  },
  components: {
    PostList
  }
}
</script>
