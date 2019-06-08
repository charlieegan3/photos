<template>
  <div>
    <div class="year" v-for="year in years">
      <router-link class="link" :to="{ name: 'archive', query: { slug: year[0][0].substring(0,4), type: 'year' } }">
        <div class="yearLabel">{{ yearFormat(year[0][0]) }}</div>
      </router-link>
      <div class="month" v-for="month in year">
        <router-link class="link" :to="{ name: 'archive', query: { slug: month[0].substring(0,7), type: 'month' } }">
          <div class="monthLabel">{{ monthFormat(month[0]) }}</div>
        </router-link>
        <div class="day" v-for="day in month">
          <router-link v-if="count(day)>0" class="link" :to="{ name: 'archive', query: { slug: day, type: 'day' } }">
            <div class="dayLabel">
              {{ dayFormat(day) }}
            </div>
          </router-link>
          <div v-if="!count(day)" class="dayLabel empty">
            {{ dayFormat(day) }}
          </div>
          <div class="count" v-if="count(day)>0">{{ count(day) }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.year {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 250px));
  column-gap: 0.5rem;
  row-gap: 0.5rem;
  margin-bottom: 0.5rem;
}
.month {
  display: grid;
  grid-template-columns: repeat(7, [col-start] 1fr);
}
.day {
  position: relative;
}
.dayLabel {
  margin: 0.1rem;
  border: 1px solid #ccc;
  text-align: center;
  padding: 0.2rem 0.3rem;
}
.empty {
  opacity: 0.3;
}
.link .dayLabel:hover {
  background-color: #ccc;
}
.yearLabel {
  font-size: 3rem;
  padding-left: 0.5rem;
}
.monthLabel {
  font-weight: bold;
  padding: 0.3rem;
}

.link, .link:visited {
  text-decoration: none;
  color: black;
}
.count{
  position: absolute;
  top: -3px;
  right: -3px;
  font-size: 0.7rem;
  background-color: Tomato;
  color: white;
  padding: 1px 3px;
  border-radius: 3px;
  z-index: 100;
}
</style>

<script>
import Moment from 'moment';
import { extendMoment } from 'moment-range';
const moment = extendMoment(Moment);

export default {
  name: "calendar",
  props: ["data"],
  methods: {
    dayFormat: function(day) { return moment(day).date() },
    monthFormat: function(day) { return moment(day).format("MMM") },
    yearFormat: function(day) { return moment(day).year() },
    count: function(day) { return this.data[day] },
  },
  computed: {
    years: function() {
      var dates = Object.keys(this.data).sort();
      var range = moment.range(dates[0], dates[dates.length-1]);
      var years = [];

      for (let year of range.by("years")) {
        var y = [];
        var yearRange = moment.range(year.startOf("year").toDate(), year.endOf("year").toDate());

        for (let month of yearRange.by("months")) {
          var m = [];
          var monthRange = moment.range(month.startOf("month").toDate(), month.endOf("month").toDate());

          for (let day of monthRange.by("days")) {
            m.push(day.format("YYYY-MM-DD"))
          }

          y.push(m);
        }

        years.unshift(y);
      }

      return years;
    }
  }
}
</script>
