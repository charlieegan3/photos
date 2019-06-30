<template>
  <div>
    <Map v-if="items" :items="items" :height="500"/>
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
    axios.get("//localhost:8000/locations.json").then(({ data }) => {
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
          link: "/locations/" + data[i].id,
          title: data[i].name,
          subtitle: data[i].count + " posts",
          count: data[i].count,
          location: {
            name: data[i].name + " (" + data[i].count + " posts)",
            lat: data[i].lat,
            long: data[i].long,
          }
        })
      }
      this.items = items;
    }
  },
  components: { Grid, Map },
}
</script>
