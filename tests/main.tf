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
	name 				  			= "First"
	description			  			= "My first job description"
	build_discard_policy  			= {
		days_to_keep_builds 		= 1
		max_n_of_builds_to_keep 	= 2
		days_to_keep_artifacts 		= 3
		max_n_of_artifacts_to_keep 	= 4
	}	
	disallow_concurrent_builds		= true
	github_project 					= {
		project_url					= "https://github.com/dihedron/libjpp"
		display_name				= "A library to emulate Java in C++" 
	}
}