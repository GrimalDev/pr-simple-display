# Pull-Requests simple display

# SSE Htmx Client

This is a simple styled htmx client that uses Server-Sent Events (SSE) to
display pull-requests from a Github repository.

### Usage

- Change the url to the endpoint (currently grimaldev.local)
- Start a server that serves the client.html file

# Simple SSE endpoint

The server uses the gh cli tool to fetch the pull-requests from the github
repository given in the .env file.

### Usage

- Use the docker-compose file to start the server
