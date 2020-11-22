/****   request.js   ****/
// 导入axios
import axios from 'axios'
// 使用做消息提醒
import Message from '../components/Message/index.js'
//1. 创建新的axios实例，
const service = axios.create({
    // 公共接口
    baseURL: process.env.BASE_API,
    // 超时时间 单位是ms，这里设置了3s的超时时间
    timeout: 3 * 1000
})
// 2.请求拦截器
service.interceptors.request.use(config => {
    //发请求前做的一些处理，数据转化，配置请求头，设置token,设置loading等，根据需求去添加
    config.data = JSON.stringify(config.data); //数据转化,也可以使用qs转换
    config.headers = {
        'Content-Type': 'application/json' //配置请求头
    }
    //注意使用token的时候需要引入cookie方法或者用本地localStorage等方法，推荐js-cookie
    const token = getCookie('API_TOKEN');//这里取token之前，你肯定需要先拿到token,存一下
    if (token) {
        config.headers.token = token; //如果要求携带在请求头中
    }
    return config
}, error => {
    Promise.reject(error)
})

// 3.响应拦截器
service.interceptors.response.use(response => {
    var data=response.data
    if (data.code == 10000) {
        return data
    }
    Message.error(data.msg)
}, error => {
    Message.error("连接服务器失败")
    return Promise.resolve(error.response)
})

export function getCookie(name) {
    var arr, reg = new RegExp("(^| )" + name + "=([^;]*)(;|$)");
    arr = document.cookie.match(reg)
    if (arr)
        return (arr[2]);
    else
        return null;
}

export function setCookie(c_name, value, expiredays) {
    var exdate = new Date();
    exdate.setDate(exdate.getDate() + expiredays);
    document.cookie = c_name + "=" + escape(value) + ((expiredays == null) ? "" : ";expires=" + exdate.toGMTString());
}

//删除cookie
export function delCookie(name) {
    var exp = new Date();
    exp.setTime(exp.getTime() - 1);
    var cval = getCookie(name);
    if (cval != null)
        document.cookie = name + "=" + cval + ";expires=" + exp.toGMTString();
}

//返回一个Promise(发送post请求)
export function fetch(method, args, version) {
    return new Promise((resolve, reject) => {
        var params = {
            method: method,
            args: args,
            version: version,
        }
        service.post("/api/gateway", params)
            .then(response => {
                resolve(response);
            }, err => {

                reject(err);
            })
            .catch((error) => {
                reject(error)
            });
    });
}