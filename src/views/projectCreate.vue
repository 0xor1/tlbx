<template>
  <div class="root">
    <h1>Project Create</h1>
    <input v-model="name" placeholder="name" @blur="validate" @keydown.enter="create">
    <span v-if="nameErr.length > 0" class="err">{{nameErr}}</span>
    <span>
      <input type="checkbox" v-model="isPublic" placeholder="isPublic" @keydown.enter="create">
      <label for="checkbox"> public</label>
    </span>
    <span>
      <select v-model="currencyCode">
        <option v-for="currency in currencies" v-bind:value="currency" v-bind:key="currency">
          {{currency}}
        </option>
      </select>
      <label> currency code</label>
    </span>
    <input v-model.number="hoursPerDay" :max="24" type="number" placeholder="hours per day" @blur="validate" @keydown.enter="create">
    <input v-model.number="daysPerWeek" :max="7" type="number" placeholder="days per week" @blur="validate" @keydown.enter="create">
    <datepicker v-model="startOn" placeholder="start on" @closed="validate"></datepicker>
    <datepicker v-model="endOn" placeholder="end on" @closed="validate"></datepicker>
    <button @click="create">create</button>
    <button @click="$router.push('/projects')">cancel</button>
    <span v-if="createErr.length > 0" class="err">{{createErr}}</span>
  </div>
</template>

<script>
  import datepicker from 'vuejs-datepicker';
  export default {
    name: 'projectCreate',
    components: {datepicker},
    data: function() {
      return {
        name: "",
        nameErr: true,
        isPublic: false,
        currencyCode: "USD",
        hoursPerDay: 8,
        daysPerWeek: 5,
        startOn: null,
        endOn: null,
        createErr: "",
        currencies: [
          "AED",
          "AFN",
          "ALL",
          "AMD",
          "ANG",
          "AOA",
          "ARS",
          "AUD",
          "AWG",
          "AZN",
          "BAM",
          "BBD",
          "BDT",
          "BGN",
          "BHD",
          "BIF",
          "BMD",
          "BND",
          "BOB",
          "BOV",
          "BRL",
          "BSD",
          "BTN",
          "BWP",
          "BYN",
          "BZD",
          "CAD",
          "CDF",
          "CHE",
          "CHF",
          "CHW",
          "CLF",
          "CLP",
          "CNY",
          "COP",
          "COU",
          "CRC",
          "CUC",
          "CUP",
          "CVE",
          "CZK",
          "DJF",
          "DKK",
          "DOP",
          "DZD",
          "EGP",
          "ERN",
          "ETB",
          "EUR",
          "FJD",
          "FKP",
          "GBP",
          "GEL",
          "GHS",
          "GIP",
          "GMD",
          "GNF",
          "GTQ",
          "GYD",
          "HKD",
          "HNL",
          "HRK",
          "HTG",
          "HUF",
          "IDR",
          "ILS",
          "INR",
          "IQD",
          "IRR",
          "ISK",
          "JMD",
          "JOD",
          "JPY",
          "KES",
          "KGS",
          "KHR",
          "KMF",
          "KPW",
          "KRW",
          "KWD",
          "KYD",
          "KZT",
          "LAK",
          "LBP",
          "LKR",
          "LRD",
          "LSL",
          "LYD",
          "MAD",
          "MDL",
          "MGA",
          "MKD",
          "MMK",
          "MNT",
          "MOP",
          "MRU",
          "MUR",
          "MVR",
          "MWK",
          "MXN",
          "MXV",
          "MYR",
          "MZN",
          "NAD",
          "NGN",
          "NIO",
          "NOK",
          "NPR",
          "NZD",
          "OMR",
          "PAB",
          "PEN",
          "PGK",
          "PHP",
          "PKR",
          "PLN",
          "PYG",
          "QAR",
          "RON",
          "RSD",
          "RUB",
          "RWF",
          "SAR",
          "SBD",
          "SCR",
          "SDG",
          "SEK",
          "SGD",
          "SHP",
          "SLL",
          "SOS",
          "SRD",
          "SSP",
          "STN",
          "SVC",
          "SYP",
          "SZL",
          "THB",
          "TJS",
          "TMT",
          "TND",
          "TOP",
          "TRY",
          "TTD",
          "TWD",
          "TZS",
          "UAH",
          "UGX",
          "USD",
          "USN",
          "UYI",
          "UYU",
          "UYW",
          "UZS",
          "VES",
          "VND",
          "VUV",
          "WST",
          "XAF",
          "XAG",
          "XAU",
          "XBA",
          "XBB",
          "XBC",
          "XBD",
          "XCD",
          "XDR",
          "XOF",
          "XPD",
          "XPF",
          "XPT",
          "XSU",
          "XTS",
          "XUA",
          "XXX",
          "YER",
          "ZAR",
          "ZMW",
          "ZWL"
        ]
      }
    },
    methods: {
      validate: function(){
        if (this.name.length > 250) {
            this.nameErr = "name must be less than 250 characters long"
        } else {
            this.nameErr = ""
        }
        if (this.hoursPerDay != null) { 
          if (this.hoursPerDay > 24) {
            this.hoursPerDay = 24
          }
          if (this.hoursPerDay < 1) {
            this.hoursPerDay = 1
          }
        }
        if (this.daysPerWeek != null) { 
          if (this.daysPerWeek > 7) {
            this.daysPerWeek = 7
          }
          if (this.daysPerWeek < 1) {
            this.daysPerWeek = 1
          }
        }
        if (this.startOn != null) {
            this.startOn.setHours(0, 0, 0, 0)
        }
        if (this.endOn != null) {
            this.endOn.setHours(0, 0, 0, 0)
        }
        if (this.startOn != null && 
          this.endOn != null &&
          this.startOn.getTime() >= this.endOn.getTime()) {
            this.endOn.setDate(this.startOn.getDate()+1)
        }
        return this.nameErr.length === 0
      },
      create: function(){
        if (this.validate()) {
          this.$api.project.create(this.name, this.isPublic, this.currencyCode, this.hoursPerDay, this.daysPerWeek, this.startOn, this.endOn).then((p)=>{
            this.$router.push('/host/'+p.host+'/project/'+p.id+'/task/'+p.id)
          })
        }
      }
    }
  }
</script>

<style scoped lang="scss">
div.root {
  & > * {
    display: block;
    margin-bottom: 5px;
  }
  button, a{
    display: inline;
    margin-right: 15px;
  }
}
.err{
  color: #c33;
}
</style>