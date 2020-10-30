
import BAPI from './bapi'


function AuthLogin(username,password,test){

	let tmpUrl = "/Auth/Login";

	
		
		tmpUrl = BAPI.AppendParam(tmpUrl, 'username', username) 
		
		tmpUrl = BAPI.AppendParam(tmpUrl, 'password', password) 
		
		tmpUrl = BAPI.AppendParam(tmpUrl, 'test', test) 
			
		
	
	return BAPI.Xhr({
		url: tmpUrl,
		method: 'get',
	
	}).then((data) => {
		return data
	})

}

function AuthLogin(username,password,test){

	let tmpUrl = "/Auth/Login";

	
		let inparam={
		
			"username":username,
		
			"password":password,
		
			"test":test,
		
		}
		
	
	return BAPI.Xhr({
		url: tmpUrl,
		method: 'post',
	
		body:inparam,
	
	}).then((data) => {
		return data
	})

}

function TokenUpload(){

	let tmpUrl = "/Token/Upload";

	
			
		
	
	return BAPI.Xhr({
		url: tmpUrl,
		method: 'get',
	
	}).then((data) => {
		return data
	})

}

function TokenLogin(username,password,test){

	let tmpUrl = "/Token/Login";

	
		let inparam={
		
			"username":username,
		
			"password":password,
		
			"test":test,
		
		}
		
	
	return BAPI.Xhr({
		url: tmpUrl,
		method: 'post',
	
		body:inparam,
	
	}).then((data) => {
		return data
	})

}

function TokenGet(key){

	let tmpUrl = "/Token/Get";

	
		
		tmpUrl = BAPI.AppendParam(tmpUrl, 'key', key) 
			
		
	
	return BAPI.Xhr({
		url: tmpUrl,
		method: 'get',
	
	}).then((data) => {
		return data
	})

}

function AuthLogin(username,password){

	let tmpUrl = "/Service/Auth";

	
		let inparam={
		
			"username":username,
		
			"password":password,
		
		}
		
	
	return BAPI.Xhr({
		url: tmpUrl,
		method: 'post',
	
		body:inparam,
	
	}).then((data) => {
		return data
	})

}
	


export{ AuthLogin }

export{ AuthLogin }

export{ TokenUpload }

export{ TokenLogin }

export{ TokenGet }

export{ ServiceAuth }
	
