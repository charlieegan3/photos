<template>
  <div id="app">
    <div id="nav">
      <router-link to="/">Home</router-link> |
      <router-link to="/tags">Tags</router-link> |
      <router-link to="/locations">Locations</router-link> |
      <router-link to="/calendar">Calendar</router-link>
    </div>
    <router-view/>
  </div>
</template>

<style>
#app {
  font-family: Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
</style>
