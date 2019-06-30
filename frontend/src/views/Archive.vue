<template>
  <div>
    <h1>{{ title }}</h1>
    <div>
      <router-link class="archive-link" v-for="archive in relatedArchives" v-bind:key="archive.slug+archive.type" :to="{ name: 'archive', params: { id: archive.slug, type: archive.type } }">{{ archive.text }}</router-link>
    </div>
    <Map v-if="items" :items="items" :height="500"/>
    <Grid v-if="items" :items="items"/>
  </div>
</template>

<script>
import axios from 'axios';
import Moment from 'moment';
import Grid from '@/components/Grid.vue'
import Map from '@/components/Map.vue'

export default {
  name: 'archive',
  created() {
    axios.get("//localhost:8000/index.json").then(({ data }) => {
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
    data: function() { this.setItems(this.$route.params.id, this.$route.params.type) },
  },
  beforeRouteUpdate: function(to, from, next) {
    this.setItems(to.params.id, to.params.type)
    next()
  },
  methods: {
    setItems: function(id, type) {
      if (!this.data) { return }
      this.items = [];
      for (var i = 0; i < this.data.length; i++) {
        if (this.includePost(this.data[i].id, id, type)) {
          this.items.push({
            post_id: this.data[i].id,
            link: "/posts/" + this.data[i].id,
            location: {
              name: this.data[i].location_name,
              lat: this.data[i].lat,
              long: this.data[i].long,
            }
          })
        }
      }
    },
    includePost: function(id, archiveID, type) {
      var slug = archiveID;
      switch (type) {
        case "year":
          if (id.substring(0,4) == slug) {
            return true
          }
          break;
        case "month":
          if (id.substring(0,7) == slug) {
            return true
          }
          break;
        case "day":
          if (id.substring(0,10) == slug) {
            return true
          }
          break;
        case "all-month":
          var date = Moment(id.substring(0,10));
          if (date.format("MMMM") == slug) {
            return true
          }
          break;
        case "month-day":
          var date = Moment(id.substring(0,10));
          if (date.format("MM-DD") == slug) {
            return true
          }
          break;
        case "weekday":
          var date = Moment(id.substring(0,10));
          if (date.format("dddd") == slug) {
            return true
          }
          break;
      }
      return false
    }
  },
  computed: {
    title: function() {
      var id = this.$route.params.id;
      var type = this.$route.params.type;
      if (type == "year") {
        return id.substring(0,4);
      }
      if (type == "month") {
        return Moment(id+"-01").format("MMMM YYYY")
      }
      if (type == "day") {
        return Moment(id).format("MMMM Do YYYY")
      }
      if (type == "all-month") {
        return id + " from all years"
      }
      if (type == "weekday") {
        return "Taken on a " + id
      }
      if (type == "month-day") {
        return "Taken on a " + Moment("2019-" + id).format("MMMM Do")
      }
    },
    relatedArchives: function() {
      var slug = this.$route.params.id;
      var type = this.$route.params.type;
      var archives = [];
      if (type == "year") {
          var years = [];
          for (var i = 0; i< this.data.length; i++) {
            var year = this.data[i].id.substring(0,4)
              if (!years.includes(year) && slug != year) {
                  years.push(year)
                  archives.push({ slug: year, type: "year", text: "Year of " + year })
              }
          }
      }
      if (type == "month") {
        var year = slug.substring(0,4);
        archives.push({ slug: year, type: "year", text: "Year of " + year })
        var date = Moment(slug + "-01");
        archives.push({ slug: date.format("MMMM"), type: "all-month", text: "All " + date.format("MMMM") })
      }
      if (type == "day") {
        var date = Moment(slug);
        archives.push({ slug: date.format("dddd"), type: "weekday", text: "Taken on a " + date.format("dddd") })
        archives.push({ slug: date.format("MM-DD"), type: "month-day", text: "Taken on a " + date.format("MMMM Do") })
        archives.push({ slug: date.format("YYYY"), type: "year", text: "Taken in " + date.format("YYYY") })
      }
      if (type == "all-month") {
        var months = ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"];
        for (var i = 0; i < months.length; i++) {
          if (months[i] != slug) {
            archives.push({ slug: months[i], type: "all-month", text: "All " + months[i] })
          }
        }
      }
      if (type == "weekday") {
        var days = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"];
        for (var i = 0; i < days.length; i++) {
          if (days[i] != slug) {
            archives.push({ slug: days[i], type: "weekday", text: "All " + days[i] + "s" })
          }
        }
      }
      return archives;
    }
  },
  components: { Grid, Map },
}
</script>

<style>
.archive-link {
  margin-right: 0.3rem;
  color: black;
}
</style>
