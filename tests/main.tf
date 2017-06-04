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
	display_name					= "This is the First project display name"
	disabled						= true		
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
	throttle_builds					= {
		rate						= 2
		period						= "day"
	}
	build_after						= {
		projects					= "pipeline-archetype"
		threshold					= "success"
	}
	periodic_build_schedule			= <<EOF
# every fifteen minutes (perhaps at :07, :22, :37, :52)
H/15 * * * *
# every ten minutes in the first half of every hour (three times, perhaps at :04, :14, :24)
H(0-29)/10 * * * *
EOF
	github_hook_trigger				= true
	scm_poll_trigger				= {
		schedule					= <<EOF
# once every two hours at 45 minutes past the hour starting at 9:45 AM and finishing at 3:45 PM every weekday.
H 9-16/2 * * 1-5
# once in every two hours slot between 9 AM and 5 PM every weekday (perhaps at 10:38 AM, 12:38 PM, 2:38 PM, 4:38 PM)
H H(9-16)/2 * * 1-5
EOF
		ignore_postcommit_hooks		= true
	}
	quiet_period					= 10
	remote_trigger_token			= "ABCDEFGHJIKLMNOPQRSTUVWXYZ"
}
