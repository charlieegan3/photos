<template>
  <div class="image" :style="'background-image: url(' + url + ');'"></div>
</template>

<script>
export default {
  props: ["post"],
  computed: {
    url: function() {
      return "http://storage.googleapis.com/charlieegan3-instagram-archive/current/" + this.post + ".jpg"
    }
  }
}
</script>

<style>
.image {
  width: 100%;
  height: 100%;
  background-position: center center;
  background-repeat: no-repeat;
  background-size: cover;
}
</style>
