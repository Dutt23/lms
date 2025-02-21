This is a library management system applications written with golang.
Using Gin (web framework), gorm (ORM) 

It uses PASETO token for authentication on certain routes.
Dragonflydb is used as cache. 
run make start_cache to download/run dragonflydb (provided docker is installed)

Redis for queuing tasks  

Swagger is used for documentation. 
run make swagger to generate the documentation.
All the documentation of the api's and models is generated inside the docs folder.
URL: http://localhost:9001/swagger/index.html

SQLITE is used as a database.

To run the application run "make server"