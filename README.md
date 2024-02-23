# Neon CRM Registration Kiosk ~ Server Backend

This project serves as the backend for a kiosk that serves POS systems running registration for events managed by [Neon CRM](https://neonone.com/). Its features include:
* multi-event storage and tracking
* identification expiry and DOB validation
* automated Neon event registration updates

...and more.

## Setup

If you'd like to build a binary, you should be able to checkout the repo and run
```
go build ...
```
in the main folder. Otherwise, if you have Go set up and working, just run
```
go run cmd/server/main.go
```
to install dependencies and start the server. 

This project uses [Viper](https://github.com/spf13/viper) to manage environment variables that it requires to run properly. 
It expects that you have a `cobra.yaml` file defined in the project directory (or wherever you have Viper set to look for config files)
with the following variables:
* `orgId` ~ the "username" for your Neon requests
* `neonkey` ~ the "password" for your Neon requests

More information can be found in the [API v2](https://developer.neoncrm.com/api-v2/) docs.

## Usage

The server, when run, listens on port 3000 for incoming HTTP requests related to the registration handling. The endpoints are defined in [root.go](cmd/server/root.go), which link to response functions found in [main.go](cmd/server/main.go), where more verbose documentation can be found. They are as follows:
* `/refresh` ~ refreshes the event listings to be chosen from when selecting events to perform registration for
* `/serverStatus` ~ obtains the general server status, with information about currently listed events and tracked attendance
* `/addEvent` ~ adds the specified event to the internal tracked events to be utilized for registration
* `/verify` ~ verifies identification validity and event registration status

Any front-end POS system can (and should) make use of these endpoints to both update its display state for registration workers, as well as confirm registration status of attendees that attempt to check in. 