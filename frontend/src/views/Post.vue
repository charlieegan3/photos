<template>
  <div>
    <div class="post-grid">
      <div>
        <div class="post-container">
          <VideoItem v-if="data.is_video" :post="this.$route.params.id"/>
          <PhotoItem v-else :post="this.$route.params.id"/>
        </div>
      </div>
      <div>
        <p>{{ caption }}</p>
        <p v-if="data">View this post on <a :href="'https://instagram.com/p/'+data.code">Instagram</a></p>
        <router-link v-if="data" :to="{ name: 'archive', params: { id: date.format('YYYY-MM-DD'), type: 'day' } }">View posts from {{ date.format("dddd Do MMMM, YYYY") }}</router-link>
        <Map v-if="mapItems" :items="mapItems" :height="200" :maxZoom="12" />
        <div v-if="index">
          <router-link v-if="nextPost" :to="'/posts/' + nextPost">Next Post</router-link>
          <router-link v-if="prevPost" :to="'/posts/' + prevPost">Prev Post</router-link>
        </div>
        <p>
          <router-link v-if="data" v-for="tag in tagList" v-bind:key="tag" :to="'/tags/' + tag">#{{ tag }}</router-link>
        </p>
        <div v-if="locationData">
          <p>{{ locationData.posts.length - 1 }} nearby posts from {{ locationData.name.replace(/,.*$/, "") }}</p>
          <Grid :items="sameLocationItems"/>
          <ul>
              <li v-for="nearby in locationData.locations">{{ nearby.name.replace(/,.*$/, "") }}</li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';
import Moment from 'moment';
import PhotoItem from '@/components/PhotoItem.vue'
import VideoItem from '@/components/VideoItem.vue'
import Map from '@/components/Map.vue'
import Grid from '@/components/Grid.vue'

export default {
  name: 'home',
  created() {
    axios.get("//localhost:8000/posts/" + this.$route.params.id + ".json").then(({ data }) => {
      this.data = data;
      this.mapItems = [
        {
          post_id: data.FullID,
          location: {
            name: data.location.name, lat: data.lat, long: data.long,
          }
        }
      ];
      axios.get("//localhost:8000/locations/" + this.data.location.id + ".json").then(({ data }) => {
        this.locationData = data;
      }).catch(function (error) { console.log(error); })
    }).catch(function (error) { console.log(error); })

    axios.get("//localhost:8000/index.json").then(({ data }) => {
      this.index = data;
    }).catch(function (error) { console.log(error); })
  },
  components: { VideoItem, PhotoItem, Map, Grid },
  data() {
    return {
      data: false,
      locationData: false,
      index: false,
      mapItems: false
    }
  },
  computed: {
    caption: function() {
      if (this.data){
        return this.data.caption
          .replace(/#[#\w\s]*$/, "")
          .replace(".\n.\n.\n", "");
      }
    },
    date: function() {
      if (this.data) {
        return Moment.unix(this.data.timestamp)
      }
    },
    tagList: function() {
      if (this.data) {
        return this.data.tags.map(function(tag) { return tag.replace("#", "")})
      }
    },
    sameLocationItems: function() {
      if (this.locationData) {
        var fullID = this.data.FullID;
        return this.locationData.posts.map(function (post) {
          if (post.FullID === fullID) { return }
          return { post_id: post.FullID, link: "/posts/" + post.FullID }
        }).filter(function (el) { return el != null; }).slice(0, 3);
      }
    },
    nextPost: function() {
      if (this.index && this.data) {
        for (var i = 0; i<this.index.length; i++) {
          if (this.index[i].id == this.data.FullID) {
            if (i == 0) {
              return null
            } else {
              return this.index[i-1].id
            }
          }
        }
      }
      return null
    },
    prevPost: function() {
      if (this.index && this.data) {
        for (var i = 0; i<this.index.length; i++) {
          if (this.index[i].id == this.data.FullID) {
            if (i == this.index.length-1) {
              return null
            } else {
              return this.index[i+1].id
            }
          }
        }
      }
      return null
    }
  }
}
</script>

<style>
.post-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
}
.post-container {
  max-width: 75vw;
  margin: 0 auto;
}
</style>
