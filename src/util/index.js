let kb = 1000
let mb = kb * kb
let gb = mb * kb
let tb = gb * kb
let pb = tb * kb
let eb = pb * tb

let self = null

function nullOr(val){
    if (val == null) {
        return null
    }
    return val
}

export default {
    install(vue){
        vue.prototype.$u = {
            _main_init_utils(vue){
                self = vue
            },
            nullOr: nullOr,
            rtr: {
                goHome (){
                    let path = "/login"
                    self.$api.user.me().then((me)=>{
                        if (me != null) {
                            path = `/host/${me.id}/projects`
                        }
                    }).catch(()=>{}).finally(()=>{
                        self.$u.rtr.goto(path)
                    })
                },
                goto (path){
                    if (self.$router.currentRoute.path != path) {
                        self.$router.push(path)
                    }
                },
                name(){
                    return nullOr(self.$router.currentRoute.name)
                },
                host(){
                    return nullOr(self.$router.currentRoute.params.host)
                },
                project(){
                    return nullOr(self.$router.currentRoute.params.project)
                },
                task(){
                    return nullOr(self.$router.currentRoute.params.task)
                }
            },
            fmt: {
                role(r) {
                    switch (r) {
                        case 0:
                            return 'admin'
                        case 1:
                            return 'writer'
                        case 2:
                            return 'reader'
                        default:
                            return 'unkown'
                    }
                },
                date(dt) {
                    if (dt == null) {
                        return ""
                    }
                    return self.$dayjs(dt).format('YYYY-MM-DD')
                },
                duration(minutes, hoursPerDay, daysPerWeek) {
                    // hoursPerDay and daysPerWeek are optional, if they arent passed
                    // or are passed as zero just show hours and minutes
                    let h = Math.floor(minutes / 60)
                    let m = minutes % 60
                    if (hoursPerDay == null || hoursPerDay === 0) {
                        if (h > 0) {
                            return h + "h " + m + "m"
                        }
                        return m + "m"
                    }
                    let d = Math.floor(h / hoursPerDay)
                    h = h % hoursPerDay
                    if (daysPerWeek == null || daysPerWeek == 0) {
                        return d + "d " + h + "h " + m + "m"
                    }
                    let w = Math.floor(d / daysPerWeek)
                    d = d % daysPerWeek
                    return w + "w " + d + "d " + h + "h " + m + "m"
                },
                cost(currencyCode, value) {
                    let symbol = currencyCode
                    // only support symbols for the major currencies
                    switch(currencyCode) {
                        case "USD":
                            symbol= '$'
                            break;
                        case "EUR":
                            symbol= '€'
                            break;
                        case "CAD":
                            symbol= 'C$'
                            break;
                        case "AUD":
                            symbol= 'A$'
                            break;
                        case "JPY":
                            symbol= '¥'
                            break;
                        case "GBP":
                            symbol= '£'
                            break;
                        case "CNY", "CNH":
                            symbol= 'CN¥'
                            break;
                        case "CHF":
                            symbol= 'Fr'
                            break;
                        case "NZD":
                            symbol= 'NZ$'
                            break;
                    }
                    return symbol + (value/100).toFixed(2)
                },
                bytes(size) {
                    let unit = "B"
                    let div = 1
                    if (size > kb) {
                        if (size < mb) {
                            unit = "KB"
                            div = kb
                        } else if (size < gb) {
                            unit = "MB"
                            div = mb
                        } else if (size < tb) {
                            unit = "GB"
                            div = gb
                        } else if (size < pb) {
                            unit = "TB"
                            div = tb
                        } else if (size < eb) {
                            unit = "PB"
                            div = pb
                        } else {
                            unit = "EB"
                            div = eb
                        }
                    }
                    if (div == 1) {
                        return (size / div) + unit
                    }
                    return (size / div).toPrecision(3) + unit
                }
            }
        }
    }
}