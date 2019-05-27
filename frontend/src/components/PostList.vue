<template>
  <div>
    <div v-for="(item, $index) in list" :key="$index">
      {{ item }}
    </div>
    <infinite-loading @infinite="infiniteHandler" spinner="spiral">
      <div slot="no-more"></div>
      <div slot="no-results"></div>
    </infinite-loading>
  </div>
</template>

<script>
import InfiniteLoading from 'vue-infinite-loading';

export default {
  props: ["posts"],
  data() {
    return {
      list: [],
      remainingPosts: this.posts,
    };
  },
  components: {
    InfiniteLoading,
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
</style>
