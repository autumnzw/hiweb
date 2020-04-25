
import BAPI from './bapi'


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
	


export{ TokenLogin }

export{ AuthLogin }
	
