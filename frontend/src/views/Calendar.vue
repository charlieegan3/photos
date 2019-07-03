<template>
  <Calendar class="ph2" v-if="data" :data="data"/>
</template>

<script>
import Moment from 'moment';
import { extendMoment } from 'moment-range';
const moment = extendMoment(Moment);
import axios from 'axios';

import Calendar from '@/components/Calendar.vue'

export default {
  name: 'calendar',
  created() {
    axios.get("//localhost:8000/calendar.json").then(({ data }) => {
      this.data = data;
    }).catch(function (error) {
      console.log(error);
    })
  },
  data() {
    return {
      data: false,
    }
  },
  components: {
    Calendar
  }
}
</script>
