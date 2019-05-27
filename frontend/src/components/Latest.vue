<template>
  <div>
    <h1>Latest</h1>
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
import axios from 'axios';
import InfiniteLoading from 'vue-infinite-loading';

export default {
  data() {
    return {
      list: [],
      data: null,
    };
  },
  components: {
    InfiniteLoading,
  },
  methods: {
    infiniteHandler($state) {
      if (this.data == null) {
        axios.get("//localhost:8000/index.json").then(({ data }) => {
          this.data = data;
          $state.loaded();
        });
      } else {
        if (this.data.length > 0) {
          this.list.push(...this.data.slice(0,18));
          this.data = this.data.slice(18);
          $state.loaded();
        } else {
          $state.complete();
        }
      }
    },
  },
}
</script>

<style scoped>
</style>
