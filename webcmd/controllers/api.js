
import axios from 'axios'
//切换路由实现进度条
import NProgress from 'nprogress'
// import 'nprogress/nprogress.css'

let http = axios.create({
  withCredentials: false,
  headers: {
      'Content-Type': 'application/x-www-form-urlencoded;charset=utf-8'
  },
});

http.defaults.withCredentials = false;//跨域安全策略

// 请求拦截器
http.interceptors.request.use(
  config => {
    let token = GetLocalToken();
    if (token != "" && config.url.indexOf("/Token?") == -1) {
      config.headers.Authorization = token;
      NProgress.start() 
    }
  
    return config;
  },
  err => {
      return Promise.reject(err);
  }
);

let baseUrl="http://localhost:8080"
function InitUrl(url){
  if(url.endWidth("/")){
    baseUrl=url
  }else{
    baseUrl=url+"/"
  }
  console.log('url:',baseUrl)
}
function Xhr( {url, body = null, method = 'get'} ) {
  if(method == "get"){
    return http
      .get(baseUrl+url)
      .then(response => response.data)
      .catch(handleError);
  }else if(method=="post"){
    return http
      .post(baseUrl+url,body)
      .then(response => response.data)
      .catch(handleError);
  }else{
    throw new TypeError("not support method", method)
  }
}

var isLogin = false;
function SetLocalToken(token){
  let tokenStr = JSON.stringify(token);
  window.localStorage.setItem('Token', tokenStr)
  isLogin = true;
}

function IsLogin(){
  return isLogin;
}

function GetLocalToken(){
  //   取出刚才存储在本地的sid值赋值给localsid
  const token = window.localStorage.getItem('Token')
  if(token) {
    try{
      const value=JSON.parse(token)
      const expires = new Date(value.profile.expires_at * 1000)
      const now = new Date().getTime();
      const isExpire = now > expires.getTime();
      if (isExpire){
          return ""
      }else{
          return value.token_type + " " + value.access_token;
      }
    }catch{
      return "";
    }
    
  }
  return ""
}

function GetUserName(){
	const token = window.localStorage.getItem('Token')
	if(token) {
		try{
			const value=JSON.parse(token)
			return value.user_name
		}catch{
			return "";
		}
	}
	return ""
}

function AppendParam(url, name, value) {
  if (url && name) {
      name += '=';
      if (url.indexOf(name) === -1) {
          if (url.indexOf('?') !== -1) {
              url += '&';
          } else {
              url += '?';
          }
          url += name + encodeURIComponent(value);
      }
  }
  return url;
}
export {SetLocalToken,GetLocalToken,GetUserName,IsLogin,AppendParam,Xhr,InitUrl}

function handleError(error) {
  let errMsg;
  if (error instanceof Response) {
      if (error.status === 405) {
          return Promise.reject(errMsg);
      }
      const body = error.json() || '';
      const err = body.error || JSON.stringify(body);
      errMsg = "${error.status} - ${error.statusText || ''} "+err;
  } else {
      errMsg = error.message ? error.message : error.toString();
  }
  // console.error(errMsg);
  return Promise.reject(errMsg);
}


function TokenGenToken(username,password){

	let tmpUrl = "/TokenGenToken";

	
	tmpUrl = AppendParam(tmpUrl, 'username', username) 
	
	tmpUrl = AppendParam(tmpUrl, 'password', password) 
		

	return Xhr({
		url: tmpUrl,
		method: 'post',
	}).then((data) => {
		return data
	})

}
	


export{ TokenGenToken }
	
