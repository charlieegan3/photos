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
    axios.get("/data/index.json").then(({ data }) => {
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
		this.items = [];
		for (var i = 0; i < data.length; i++) {
			this.items.push({
				post_id: data[i].id,
				link: "/posts/" + data[i].id,
			})
		}
    }
  },
  components: {
    Grid
  }
}
</script>
