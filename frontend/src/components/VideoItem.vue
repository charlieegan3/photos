<template>
  <video class="w-100" style autoplay loop muted>
    <source type="video/mp4" :src="url">
    Video not supported
  </video>
</template>

<script>
export default {
  props: ["post"],
  computed: {
    url: function() {
      return "https://storage.googleapis.com/charlieegan3-photos/current/" + this.post + ".mp4"
    }
  }
}
</script>

<style>
</style>
