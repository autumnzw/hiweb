
import BAPI from './bapi'


function TokenGenToken(username,password){

	let tmpUrl = "/Token/GenToken";

	
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
	


export{ TokenGenToken }
	
