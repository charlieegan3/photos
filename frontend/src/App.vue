<template>
  <div id="app">
    <div class="sans-serif display mw7 mw9-ns center cf tc bg-near-black f6 f5-ns">
      <router-link class="fl f4 w-100 w-30-l pt1 pb2 db dim link gray" to="/">
        <strong>
          <span class="moon-gray">charlieegan</span><span class="green">3
          </span>&nbsp;<span class="blue">photos</span>
        </strong>
      </router-link>
      <router-link class="fl w-50 w-25-m w-10-l pt1 pb2 db dim link gray" to="/">Latest</router-link>
      <router-link class="fl w-50 w-25-m w-10-l pt1 pb2 db dim link gray" to="/tags">Tags</router-link>
      <router-link class="fl w-50 w-25-m w-10-l pt1 pb2 db dim link gray" to="/locations">Locations</router-link>
      <router-link class="fl w-50 w-25-m w-10-l pt1 pb2 db dim link gray" to="/calendar">Calendar</router-link>
    </div>
    <router-view class="mw8 center pa3-ns ph5-l" :key="$route.fullPath"></router-view>
  </div>
</template>

<style>
#app {
  font-family: "IBM Plex Serif", serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
.sans-serif {
  font-family: "Montserrat", sans-serif;
}
</style>
