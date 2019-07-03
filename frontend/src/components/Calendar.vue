<template>
  <div>
    <div class="year" v-for="year in years">
      <router-link class="link" :to="{ name: 'archive', params: { id: year[0][0].substring(0,4), type: 'year' } }">
        <div class="yearLabel gray">{{ yearFormat(year[0][0]) }}</div>
      </router-link>
      <div class="month" v-for="month in year">
        <router-link class="link" :to="{ name: 'archive', params: { id: month[0].substring(0,7), type: 'month' } }">
          <div class="monthLabel gray">{{ monthFormat(month[0]) }}</div>
        </router-link>
        <div class="day" v-for="day in month">
          <router-link v-if="count(day)>0" class="link" :to="{ name: 'archive', params: { id: day, type: 'day' } }">
            <div class="dayLabel gray">
              {{ dayFormat(day) }}
            </div>
          </router-link>
          <div v-if="!count(day)" class="dayLabel empty mid-gray">
            {{ dayFormat(day) }}
          </div>
          <div class="count" v-if="count(day)>1">{{ count(day) }}</div>
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
}
.count{
  position: absolute;
  top: -3px;
  right: -3px;
  font-size: 0.7rem;
  background-color: #facc91;
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
      var from = dates[0];
      var to = dates[dates.length-1];
      var range = moment.range(from, to);
      var years = [];

      for (let year of range.by("years")) {
        var y = [];
        var yearRange = moment.range(year.startOf("year").toDate(), year.endOf("year").toDate());

        for (let month of yearRange.by("months")) {
          if (month.isAfter(to) || month.isBefore(moment(from).startOf("month"))) {
            continue
          }
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
