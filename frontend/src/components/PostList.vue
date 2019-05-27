<template>
  <div class="list" >
    <div class="item" v-for="(item, $index) in list" :key="$index">
      <router-link :to="'/posts/' + item.id">
        <VideoItem v-if="item.is_video" :post="item.id"/>
        <PhotoItem v-else :post="item.id"/>
      </router-link>
    </div>
    <infinite-loading @infinite="infiniteHandler" spinner="spiral">
      <div slot="no-more"></div>
      <div slot="no-results"></div>
    </infinite-loading>
  </div>
</template>

<script>
import InfiniteLoading from 'vue-infinite-loading';
import PhotoItem from '@/components/PhotoItem.vue'
import VideoItem from '@/components/VideoItem.vue'

export default {
  props: ["posts"],
  data() {
    return {
      list: [],
      remainingPosts: this.posts,
    };
  },
  components: {
    InfiniteLoading, VideoItem, PhotoItem,
  },
  methods: {
    infiniteHandler($state) {
      if (this.remainingPosts.length > 0) {
        this.list.push(...this.remainingPosts.slice(0,18));
        this.remainingPosts = this.remainingPosts.slice(18);
        $state.loaded();
      } else {
        $state.complete();
      }
    },
  },
}
</script>

<style scoped>
.list {
  max-width: 900px;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  grid-auto-rows: 1fr;
}
.list::before {
  content: '';
  width: 0;
  padding-bottom: 100%;
  grid-row: 1 / 1;
  grid-column: 1 / 1;
}
.list > *:first-child {
  grid-row: 1 / 1;
  grid-column: 1 / 1;
}
</style>
