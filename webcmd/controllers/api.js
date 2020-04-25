
import BAPI from './bapi'


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

function AuthLogin(username,password){

	let tmpUrl = "/Auth/Login";

	
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
	


export{ ServiceAuth }

export{ AuthLogin }

export{ TokenLogin }

export{ TokenGet }
	
