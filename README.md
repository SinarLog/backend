# SinarLog Backend

## Introduction
This repository stores the code of SinarLog's backend application. SinarLog is an application that makes attendances easier than ever. Users are able to manage their attendances and HRs are able to view their employees' attendances. With SinarLog, applying for leave request is even better. Staff and/or managers do not need to hassle just to apply leave. This also benefits HRs on approving and disapproving a leave request.

## Getting Started
For the current version of SinarLog, we focus on three entities. A staff, manager and an HR. A staff can only have one manager and managers can have many staffs. HR on the other hand, acts as an admin. HR does not have a manager nor staffs under them. They are a stand alone unit of admin functionalities.

With SinarLog, staff and managers can manage their attendances. Clocking in and out as well as requesting a leave. HR can view employees' attendances and accept or reject a leave request. Since a staff is under a manager, managers can also accep or reject  their staffs' leave request.

## How to clone
These are the requirements to clone, run and develop this project:
- Manually
	1. Having Golang installed in your local machine.
	2. Having PostgreSQL installed in your local machine.
	3. Having Redis installed in your local machine.
	4. Having a Firebase Storage account.
	5. Having an email account for sending emails from the app.
	6. Having an .env and/or .env.development in your local (please edit this in config/config.go) file.
	7. Next, having the database setup in your local (if you want this to be automatic, edit it in pkg/postgres.go)
	8. Run `go install` then `go run main.go` to run it.

- Docker
	1. Have your environments ready, follow our env template.
	1. Then do `docker compose up` to start it.