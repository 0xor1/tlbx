<template>
  <div class="root">
    <div v-if="loading">
      loading...
    </div>
    <div v-else>
      <h1>project {{isCreate? 'create': 'update'}}</h1>
      <input v-model="name" placeholder="name" @blur="validate" @keydown.enter="ok">
      <span v-if="nameErr.length > 0" class="err">{{nameErr}}</span>
      <span>
        <input type="checkbox" v-model="isPublic" placeholder="isPublic" @keydown.enter="ok">
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
      <input v-model.number="hoursPerDay" :min="0" :max="24" type="number" placeholder="hours per day" @blur="validate" @keydown.enter="ok">
      <input v-model.number="daysPerWeek" :min="0" :max="7" type="number" placeholder="days per week" @blur="validate" @keydown.enter="ok">
      <datepicker v-model="startOn" placeholder="start on" @closed="validate"></datepicker>
      <datepicker v-model="endOn" placeholder="end on" @closed="validate"></datepicker>
      <button @click="ok">{{isCreate? 'create': 'update'}}</button>
      <button @click="cancel">cancel</button>
      <span v-if="err.length > 0" class="err">{{err}}</span>
    </div>
  </div>
</template>

<script>
  import datepicker from 'vuejs-datepicker';
  export default {
    name: 'projectCreateOrUpdate',
    components: {datepicker},
    data: function() {
      return this.initState()
    },
    computed: {
      isCreate(){
        return this.$u.rtr.name() == "projectCreate"
      },
      isUpdate(){
        return !this.isCreate
      }
    },
    methods: {
      initState (){
        return {
          loading: true,
          name: "",
          nameErr: true,
          isPublic: false,
          currencyCode: "USD",
          hoursPerDay: null,
          daysPerWeek: null,
          startOn: null,
          endOn: null,
          err: "",
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
      init(){
        for(const [key, value] of Object.entries(this.initState())) {
          this[key] = value
        }
        if (this.isUpdate) {
          this.$api.user.me().then((me)=>{
            if (me.id !== this.$u.rtr.host()) {
              this.$u.rtr.goHome()
              return
            }
            this.$api.project.one(this.$u.rtr.host(), this.$u.rtr.project()).then((p)=>{
              this.name = p.name
              this.isPublic = p.isPublic
              this.currencyCode = p.currencyCode
              this.hoursPerDay = p.hoursPerDay
              this.daysPerWeek = p.daysPerWeek
              if (p.startOn != null) {
                this.startOn = new Date(p.startOn)
              }
              if (p.endOn != null) {
                this.endOn = new Date(p.endOn)
              }
              this.loading = false
            })
          })
        } else {
          this.loading = false
        }
      },
      validate(){
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
            this.hoursPerDay = null
          }
        }
        if (this.daysPerWeek != null) { 
          if (this.daysPerWeek > 7) {
            this.daysPerWeek = 7
          }
          if (this.daysPerWeek < 1) {
            this.daysPerWeek = null
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
      ok(){
        if (this.validate()) {
          if (this.isCreate) {
            this.$api.project.create(this.name, this.isPublic, this.currencyCode, this.hoursPerDay, this.daysPerWeek, this.startOn, this.endOn).then((p)=>{
              this.$u.rtr.goto(`/host/${p.host}/project/${p.id}/task/${p.id}`)
            })
          } else {
            this.$api.project.updateOne({
              id: this.$u.rtr.project(), 
              name: {v: this.name},
              isPublic: {v: this.isPublic},
              currencyCode: {v: this.currencyCode},
              hoursPerDay: {v: this.hoursPerDay},
              daysPerWeek: {v: this.daysPerWeek},
              startOn: {v: this.startOn},
              endOn: {v: this.endOn}
            }).then((p)=>{
              this.$u.rtr.goto(`/host/${p.host}/project/${p.project}/task/${p.project}`)
            })
          }
        }
      },
      cancel(){
        this.$u.rtr.goHome()
      }
    },
    mounted(){
      this.init()
    },
    watch: {
      $route () {
        this.init()
      }
    }
  }
</script>

<style scoped lang="scss">
div.root > div {
  padding: 2.6pc 0 0 1.3pc;
  & > * {
    display: block;
    margin-bottom: 5px;
  }
  button, a{
    display: inline;
    margin-right: 15px;
  }
  input[type="number"] {
    width: 10pc;
  }
}
.err{
  color: #c33;
}
</style>