/*
 * The JenkinsCI server instance.
 */
provider "jenkins" {
	server_url          = "http://localhost:8080/"
	ca_cert 		    = false
	username 		   	= "Administrator"
	password 		    = "password"
}


/*
 * The first JenkinsCI Job.
 */ 
resource "jenkins_job" "first" {
	name 				= "First"
}
