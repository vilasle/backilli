package main

import(

)

// Short map and plans
// illi [operation] {init | run | check} [subject] [--config | -c ]  <string> [--env | -e] <string> [<Additional operation args>]  
// Operation:
//	init
//		- subject:
//			- config  (run CLI interface generation config file )
//			- env	  (run CLI interface generation env file )
//  run 
//		#TODO Don't want to do it
//		- subject:		
//			- interface - running web interface for setting and monitoring (does not work)
//				- additional_agrs:
//					--http	  - listen to port ( default:1780) 
//			- backup    - running backup
// 
//	check 
//		- subject:
//			- all ( load all setting and check access, e.g access to disk or postgres)
//illi interface --config file.conf --env file.env --http 8080
//	comment: run http server which will listen port 8080 and after finishing setting it write all setting to file.conf and file.env
//	
//illi run backup --config file.conf --env file.env 
//	comment: run backup process with using setting from file.conf and file.env
//
//illi init config --config file.conf
//	comment: run cli and generate config and save it to file.conf
//
//illi init env --env file.conf
//	comment: run cli and generate enviroment vars and save it to file.env
//

func main() {
	
}