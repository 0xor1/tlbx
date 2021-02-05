import marked from 'marked'
import dompurify from 'dompurify'

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
            cnsts: {
                time: "time",
                cost: "cost"
            },
            copyProps(src, dst) {
                for(const [key, value] of Object.entries(src)) {
                    dst[key] = value
                }
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
            perm: {
                canAdmin: (pMe) => pMe != null && pMe.isActive === true && pMe.role < 1,
                canWrite: (pMe) => pMe != null && pMe.isActive === true && pMe.role < 2,
                canRead: (pMe) => pMe != null && pMe.isActive === true && pMe.role < 3
            },
            fmt: {
                mdLinkify(txt){
                    // replace all instances of [foo](bar) with <a href="bar">foo</a> tag
                    return txt.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank">$1</a>')
                },
                md(txt) {
                    return `<div class="markdown">${dompurify.sanitize(marked(txt))}</div>`
                },
                ellipsis(txt, len) {
                    if (len == null) {
                        throw new Error('len must be defined')
                    }
                    let res = txt
                    if (txt.length > len) {
                        res = txt.substring(0, len) + '...'
                    }
                    return res
                },
                role(r) {
                    switch (r) {
                        case null:
                            return 'none'
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
                datetime(dt) {
                    if (dt == null) {
                        return ""
                    }
                    return self.$dayjs(dt).format('YYYY-MM-DD HH:mm')
                },
                time(minutes, hoursPerDay, daysPerWeek) {
                    // hoursPerDay and daysPerWeek are optional, if they arent passed
                    // or are passed as zero just show hours and minutes
                    let h = Math.floor(minutes / 60)
                    let m = minutes % 60
                    if (hoursPerDay == null || hoursPerDay === 0) {
                        let res = ""
                        if (h > 0) {
                            res += h + "h"
                        }
                        if (m > 0 || h == 0) {
                            if (res != "") {
                                res += " "
                            }
                            res += m + "m"
                        }
                        return res
                    }
                    let d = Math.floor(h / hoursPerDay)
                    h = h % hoursPerDay
                    if (daysPerWeek == null || daysPerWeek == 0) {
                        let res = ""
                        if (d > 0) {
                            res += d + "d"
                        }
                        if (h > 0) {
                            if (res != "") {
                                res += " "
                            }
                            res += h + "h"
                        }
                        if (m > 0 || (h == 0 && d == 0)) {
                            if (res != "") {
                                res += " "
                            }
                            res += m + "m"
                        }
                        return res
                    }
                    let w = Math.floor(d / daysPerWeek)
                    d = d % daysPerWeek
                    let res = ""
                    if (w > 0) {
                        res += w + "w"
                    }
                    if (d > 0) {
                        if (res != "") {
                            res += " "
                        }
                        res += d + "d"
                    }
                    if (h > 0) {
                        if (res != "") {
                            res += " "
                        }
                        res += h + "h"
                    }
                    if (m > 0 || (h == 0 && d == 0 && w == 0)) {
                        if (res != "") {
                            res += " "
                        }
                        res += m + "m"
                    }
                    return res
                },
                currencySymbol(code){
                    let symbol = code
                    if (code != null) {
                        // only support symbols for the major currencies
                        switch(code) {
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
                    }
                    return symbol
                },
                cost(value, abbreviate) {
                    let div = 1
                    let decPlaces = 2
                    let orderSymbol = ""
                    value /= 100 
                    if (abbreviate) {
                        if (value < 1000) {
                            div = 1
                        } else if (value < 10000) {
                            div = 1000
                            decPlaces = 2
                            orderSymbol = "k"
                        } else if (value < 100000) {
                            div = 1000
                            decPlaces = 1
                            orderSymbol = "k"
                        } else if (value < 1000000) {
                            div = 1000
                            decPlaces = 0
                            orderSymbol = "k"
                        } else if (value < 10000000) {
                            div = 1000000
                            decPlaces = 2
                            orderSymbol = "m"
                        } else if (value < 100000000) {
                            div = 1000000
                            decPlaces = 1
                            orderSymbol = "m"
                        } else if (value < 1000000000) {
                            div = 1000000
                            decPlaces = 0
                            orderSymbol = "m"
                        } else if (value < 10000000000) {
                            div = 1000000000
                            decPlaces = 2
                            orderSymbol = "b"
                        } else if (value < 100000000000) {
                            div = 1000000000
                            decPlaces = 1
                            orderSymbol = "b"
                        } else if (value < 1000000000000) {
                            div = 1000000000
                            decPlaces = 0
                            orderSymbol = "b"
                        } else if (value < 10000000000000) {
                            div = 1000000000000
                            decPlaces = 2
                            orderSymbol = "t"
                        } else if (value < 100000000000000) {
                            div = 1000000000000
                            decPlaces = 1
                            orderSymbol = "t"
                        } else if (value < 1000000000000000) {
                            div = 1000000000000
                            decPlaces = 0
                            orderSymbol = "t"
                        }
                    }
                    return (value/div).toFixed(decPlaces) + orderSymbol
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
            },
            parse: {
                time(str){
                    if (str != null && str.length > 0) {
                        str = str.trim()
                        if (str == "0") {
                            return 0
                        }
                        let match = str.match(/^((\d+)h)? *((\d+)m)?$/)
                        if (match != null && match[0] != null && match[0].length > 0) {
                            let newVal = null
                            if (match[2] != null) {
                                newVal += parseInt(match[2], 10) * 60
                            }
                            if (match[4] != null) {
                                newVal += parseInt(match[4], 10)
                            }
                            if (!isNaN(newVal) && newVal != null) {
                                return newVal
                            }
                        }
                    }
                    return null
                },
                cost(str) {
                    if (str != null && str.length > 0) {
                        str = str.trim()
                        let match = str.match(/^(\d*)(\.|,)?(\d{0,2})?$/)
                        if (match != null && match[0] != null && match[0].length > 0) {
                            if (match[3] == null) {
                                match[3] = "00"
                            }
                            if (match[3].length == 1) {
                                match[3] += "0"
                            }
                            let newVal = parseInt(match[1]+match[3])
                            newVal = Math.floor(newVal)
                            if (!isNaN(newVal) && newVal != null) {
                                return newVal
                            }
                        }
                    }
                    return null
                }
            }
        }
    }
}