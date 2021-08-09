<template>
  <div v-if="page == 'post'" class="image" :style="'background-image: url(' + url + '?page=post);'"></div>
  <div v-else class="image" :style="'background-image: url(' + url + ');'"></div>
</template>

<script>
export default {
  props: ["post", "page"],
  computed: {
    url: function() {
      return "https://storage.googleapis.com/charlieegan3-photos/current/" + this.post + ".jpg"
    }
  }
}
</script>

<style>
.image {
  width: 100%;
  background-position: center center;
  background-repeat: no-repeat;
  background-size: cover;
}
.image:after {
  content: "";
  display: block;
  padding-bottom: 100%;
}
</style>
