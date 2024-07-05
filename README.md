# phishing-db
Phishing DB contains 2 services that communicate with MongoDB, one serves as an API and the second one fetches the data and stores it

Fetcher:
  Fetches JSON data from "http://data.phishtank.com/data/online-valid.json", since this is a database dump it can contain a lot of data, so we are consuming this as an HTTP stream and storing data one by one. Fetcher is fetching data once per hour and, since there are no available API keys for phishtank and because of basic failures that can occur during HTTP request, we are retrying the request on failure once per minute (default is 10 times, but can be configured in config.go (note: this can be moved to an env var)). While fetching data from "phishtank", in parallel we are writing this data to a MongoDB with just one modification: we are adding submission_time_unix for future querying (note: this value must be indexed in Mongo). Update of database is done in transaction to not disrupt queries executed by API. (note: It is possible that we can speed up the process by using batch inserts and by analysing the nature of data (there can be a lot of stale data, so we may not need to update the whole collection every time))

API:
  API service is a REST API server with just two endpoints /download_report?from=(unix timestamp in seconds)&to=(optional, unix timestamp in seconds) and /search_domain?domain(string). Download report returns a large JSON file, since part of the response required computation over obtained values -- we cannot stream the response directly, which means that we are requiring a lot of memory per call (note: we can add a caching layer that will store response for already computed data). Download report endpoint require "from" query parameter and "to" is optional, it contains list of phishing urls, length of the list and TLDs with occurences in ordered list. Search domain endpoint is more memory effective (note: but can be hard on DB, since we don't URL field indexed), it returns all matches for URLs in database.

Notes:
  This is a project that is done quick, so here i will be talking more about possible future work required for this project
  - Code can use more abstractions, like abstraction over MongoDB collections, so we would be able to interchange the DB and be able to unit test part of the codes that require to interact with DB
  - Since a lot of the code is tightly linked with database, having only unit tests would not be sufficient, we will need to be sure that we are using db related libraries correctly and for this we would need some e2e tests
  - Searching for TLD in URL may need a further review, since i already have found one exception with IPV4 addresses, there can be more
  - In theory we can fail all of the HTTP requests with retries (default number of retries per hour is 10), currently "Fetcher" will retry only after one minute since i didn't want to be disrespectful to phishtank and retry right away
  - Currently default level in API service is set to DEBUG and "Fetcher" uses INFO logs for some of the logs that should be on DEBUG level
  - ...

How to start a service on your machine:
  Docker is required for running this project, execute:
  `docker-compose up` or on newer versions of docker `docker compose up` and it should start services