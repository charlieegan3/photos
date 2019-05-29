<template>
  <div class="list" >
    <div class="item" v-for="(item, $index) in list" :key="$index">
      <router-link :to="item.link" v-if="item.title">
        <div class="item-text" >
          <p class="item-title">{{ item.title }}</p>
          <p class="item-subtitle">{{ item.subtitle }}</p>
        </div>
      </router-link>
      <router-link :to="item.link">
        <VideoItem v-if="item.is_video" :post="item.post_id"/>
        <PhotoItem v-else :post="item.post_id"/>
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
  props: ["items"],
  data() {
    return {
      list: [],
      remainingItems: this.items,
    };
  },
  components: {
    InfiniteLoading, VideoItem, PhotoItem,
  },
  methods: {
    infiniteHandler($state) {
      if (this.remainingItems.length > 0) {
        this.list.push(...this.remainingItems.slice(0,18));
        this.remainingItems = this.remainingItems.slice(18);
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
.item {
  position: relative;
}
.item-text {
  background-color: rgba(0,0,0,0.5);
  z-index: 1000;
  height: 65%;
  position: absolute;
  width: 100%;
  padding-top: 35%;
  text-align: center;
}
.item-title {
  color: white;
  overflow: hidden;
  text-overflow: ellipsis;
}
.item-subtitle {
  color: rgba(255,255,255,0.6);
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
