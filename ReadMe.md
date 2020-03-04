# Bloom Credit Batch And API

A project that processes a batch of (simulated) data in a textual fixed-width format, stores it in a
relational database, and makes it discoverable via a REST API.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

What things you need to install to run this repo

[Golang](https://golang.org/)

[Docker](https://www.docker.com/products/docker-desktop)

[test.dat](https://drive.google.com/uc?export=download&confirm=mxAW&id=1WDGWcePae1Q8oFLbyBOSzwSBBIwTnLmO)

### Installing

A step by step series of examples that tell you how to get a development env running

Get the Project
```
git clone
```

Unzip tar.b2 (Be sure to copy test.dat to the batch folder)
```
tar -jf test.tar.bz2
```

Spin up the docker Postgres DB
```
docker run --name some-postgres -e POSTGRES_PASSWORD=pass -p 5432:5432 -d postgres:10.0
```

Import customer.sql
```
cd sql
docker cp ./customers.sql some-postgres:./customers.sql
docker exec -it some-postgres psql -h localhost -U postgres -f ./customers.sql
```

Start Batch Import of test.dat
```
cd batch
go build readBat.go
readBat ../test.dat
```

Start up Api Server
```
go build -o restApi main.go
restApi
```
## Querying API

In order to get a consumers id ( first_name and last_name must be provided in query string format)
```
curl -X GET 'http://localhost:7000/getId?first_name=aaron&last_name=hunt' 
```

Once you have a consumers id to get a list of credit tags (id must be provided in query string format)
```
curl -X GET 'http://localhost:7000/customer?id=979fff8c-09f2-438f-99dd-b503a871d772' 
```

In order to get stats for a credit tag (tag must be provided in query string format)
```
curl -X GET 'http://localhost:7000/stats?tag=X0001'
```

## Running the tests

Test for batch are a work in progress

To run the test for api:
```
go test
```
## Deployment

Pretty much any platform should work. The only thing you will need to do is update the code by removing the hard coded database information

## Cleanup

For turning off docker the postgres db
```
docker stop some-postgres
docker rm some-postgres
```

## Authors

* **Adefemi Adeyemi** - *Initial work* - [ImperiousEnterprise](https://github.com/ImperiousEnterprise)

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details


