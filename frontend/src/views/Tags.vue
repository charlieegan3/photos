<template>
  <div>
    <Grid v-if="items" :items="items"/>
  </div>
</template>

<script>
import Grid from '@/components/Grid.vue'
import axios from 'axios';

export default {
  name: 'home',
  created() {
    axios.get("/data/tags.json").then(({ data }) => {
      this.data = data;
    }).catch(function (error) {
      console.log(error);
    })
  },
  data() {
    return {
      data: false,
      items: false,
    }
  },
  watch: {
    data: function(data) {
      var items = [];
      for (var i = 0; i < data.length; i++) {
        items.push({
          post_id: data[i].most_recent,
          link: "/tags/" + data[i].name,
          title: "#" + data[i].name,
          subtitle: data[i].count + " posts",
        })
      }
      this.items = items;
    }
  },
  components: {
    Grid
  }
}
</script>
