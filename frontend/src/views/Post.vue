<template>
  <div>
    <div class="post-grid">
      <div>
        <div class="post-container">
          <VideoItem v-if="data.is_video" :post="this.$route.params.id"/>
          <PhotoItem v-else :post="this.$route.params.id" :page="'post'"/>
        </div>
        <div v-if="index" class="pt1">
          <router-link class="fl silver no-underline" v-if="nextPost" :to="'/posts/' + nextPost">&larr;Next</router-link>
          <router-link class="fr silver no-underline" v-if="prevPost" :to="'/posts/' + prevPost">Previous&rarr;</router-link>
        </div>
      </div>
      <div class="f7 f5-ns ph3 pt2 pt0-l gray">
        <p class="mt0 silver">{{ caption }}</p>
        <Map v-if="mapItems" :items="mapItems" :height="200" :maxZoom="12" />
        <p v-if="data">View this post on <a class="no-underline silver" :href="'https://instagram.com/p/'+data.code">Instagram</a></p>
        <p v-if="data">View all from <router-link class="silver no-underline" :to="{ name: 'archive', params: { id: date.format('YYYY-MM-DD'), type: 'day' } }">{{ date.format("dddd Do MMMM, YYYY") }}</router-link>
        <p>
          <template v-for="tag in tagList" v-if="data">
            <router-link class="silver no-underline f7 pr1" :to="'/tags/' + tag">#{{ tag }}</router-link>
            &Tab;
          </template>
        </p>
        <div v-if="locationData">
          <div v-if="locationData.locations !== null">
            <p v-if="locationData.posts.length > 0">
              {{ locationData.posts.length - 1 }} nearby posts from {{ locationData.name.replace(/,.*$/, "") }}
            </p>
            <Grid class="h4" :items="sameLocationItems"/>
            <p>
              Browse nearby
              <router-link class="di silver no-underline" :to="'/locations/' + locationData.id">
                {{ locationData.name.replace(/,.*$/, "") }}
              </router-link>
            </p>
            <ul v-if="locationData.locations !== null">
              <li v-for="nearby in locationData.locations">
                <router-link class="di silver no-underline" :to="'/locations/' + nearby.id">
                  {{ nearby.name.replace(/,.*$/, "") }}
                </router-link>
              </li>
            </ul>
          </div>
          <p v-else>
            {{ locationData.name.replace(/,.*$/, "") }}
          </p>
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
