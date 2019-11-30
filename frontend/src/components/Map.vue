<template>
  <div :style="style" id="map" data-location="location"></div>
</template>

<script>
import Mapbox from "mapbox-gl";

export default {
  props: {"height": Number, "items":{ required: true }, "maxZoom": { default: 12 }, "location": {} },

  data() {
    return {
      accessToken: "pk.eyJ1IjoiY2hhcmxpZWVnYW4zIiwiYSI6ImNqZzB2MDdxZTFjNTAyeHRsemQwbjVsZXIifQ.NC-7ANrzAfu5RDfpCQIEMg",
      mapStyle: "mapbox://styles/charlieegan3/cjg13skaf0o7c2spats0nxsfy",
      markers: [],
    }
  },

  watch: {
    items: function() {
      Mapbox.accessToken = this.accessToken;
      this.map = new Mapbox.Map({
        container: 'map',
        style: this.mapStyle,
      });
      var map = this.map;
      map.scrollZoom.disable();
      map.addControl(new Mapbox.NavigationControl());
      map.addControl(new Mapbox.FullscreenControl());

      var markers = this.markers;
      this.geoJSON.features.forEach(function(feature) {
        var el = document.createElement('div');
        el.className = 'marker';
        el.setAttribute("data-icon", feature.properties.icon);

        var marker = new Mapbox.Marker(el)
          .setLngLat(feature.geometry.coordinates)

        if (feature.properties.link) {
          marker.setPopup(new Mapbox.Popup({offset: 25}).setHTML('<a href="' + feature.properties.link + '">' + feature.properties.title + '</a>'));
        }

        marker.addTo(map);
        markers.push(marker);
      });

      this.setMarkerSize();
      map.on('zoom', this.setMarkerSize);

      var bounds = new Mapbox.LngLatBounds();
      markers.forEach(function(feature) { bounds.extend(feature.getLngLat()) });

      map.fitBounds(bounds, { padding: this.height / 4, maxZoom: this.maxZoom, duration: 100 });
    }
  },

  methods: {
  setMarkerSize: function() {
    var scale = Math.pow(this.map.getZoom(), 2) / 2;
    if (scale < 10) { scale = 10 }
    if (scale > 40) { scale = 40 }
    this.markers.forEach(function(marker) {
      var elem = marker.getElement();
      if (scale > 15) {
        elem.style.backgroundImage = "url(" + elem.getAttribute("data-icon") + ")";
      }
      elem.style.width = scale + "px";
      elem.style.height = scale + "px";
    })
  }
  },

  computed: {
    style() {
      return "height: " + this.height + "px";
    },
    geoJSON() {
      var features = [];
      for (var i = 0; i < this.items.length; i++) {
        features.push({
          "type": "Feature",
          "geometry": {
            "type": "Point",
            "coordinates": [
              this.items[i].location.long,
              this.items[i].location.lat,
            ]
          },
          "properties": {
                  "title": this.items[i].location.name,
                  "link": this.items[i].link,
                  "icon": "https://images.weserv.nl/?url=storage.googleapis.com/charlieegan3-instagram-archive/current/" + this.items[i].post_id + ".jpg&w=50",
                  "postCount":  this.items[i].location.count,
          }
        });
      }
      return {
        type: "FeatureCollection",
        features: features
      }
    }
  }
};
</script>

<style>
.marker {
  background-size: cover;
  border-radius: 50%;
  border: 1px solid black;
  cursor: pointer;
}
.mapboxgl-popup {
  max-width: 200px;
}
.mapboxgl-popup-content {
  text-align: center;
  font-family: 'Open Sans', sans-serif;
}
</style>
