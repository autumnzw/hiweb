
import BAPI from './bapi'


function AuthLogin(password,username){

	let tmpUrl = "/Auth/Login";

	
		let inparam={
		
			"password":password,
		
			"username":username,
		
		}
		
	
	return BAPI.Xhr({
		url: tmpUrl,
		method: 'post',
	
		body:inparam,
	
	}).then((data) => {
		return data
	})

}

function TokenUpload(file){

	let tmpUrl = "/Token/Upload";

	
		
		tmpUrl = BAPI.AppendParam(tmpUrl, 'file', file) 
			
		
	
	return BAPI.Xhr({
		url: tmpUrl,
		method: 'get',
	
	}).then((data) => {
		return data
	})

}

function TokenLogin(username,password){

	let tmpUrl = "/Token/Login";

	
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

function ServiceAuth(username,password){

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

export{ TokenUpload }

export{ TokenLogin }

export{ TokenGet }

export{ ServiceAuth }
	
