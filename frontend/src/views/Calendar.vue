<template>
  <div>
    <h1>Calendar</h1>
    <div class="year" v-for="year in data">
      <div class="yearLabel">{{ yearFormat(year[0][0]) }}</div>
      <div class="month" v-for="month in year">
        <div class="monthLabel">{{ monthFormat(month[0]) }}</div>
        <div class="day" v-for="day in month">
          <div class="dayLabel">{{ dayFormat(day) }}</div>
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
.dayLabel {
  margin: 0.1rem;
  border: 1px solid #ccc;
  text-align: center;
  padding: 0.2rem 0.3rem;
}
.yearLabel {
  font-size: 3rem;
  padding-left: 0.5rem;
}
.monthLabel {
  font-weight: bold;
  padding: 0.3rem;
}
</style>

<script>
import Moment from 'moment';
import { extendMoment } from 'moment-range';
const moment = extendMoment(Moment);

export default {
  name: 'calendar',
  data() {
    return {
      startDate: new Date(2013, 2, 24),
      endDate: new Date(2019, 7, 3),
    }
  },
  methods: {
    dayFormat: function(day) { return moment(day).date() },
    monthFormat: function(day) { return moment(day).format("MMM") },
    yearFormat: function(day) { return moment(day).year() },
  },
  computed: {
    data: function() {
      var range = moment.range(this.startDate, this.endDate);
      var data = [];

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

        data.unshift(y);
      }

      return data;
    }
  }
}
</script>
