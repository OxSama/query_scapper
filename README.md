# query_scapper



## Prerequisites

Before you begin, ensure you have the following installed:

- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/install/)
- (Optional) MySQL and PostgreSQL CLI tools for database access.



## Setup Instructions

1. **Clone the Repository**:
   ```bash
   git clone <repository-url>
   cd sql_scrapper
   ```

2. **Build and Run the Application**: 

Run the following command to build and start the Docker containers in the background:

    ```bash
        docker compose up -d --build
    ```

3. **Verify Running Containers**: 

    ```bash
        docker ps
    ```


App containers should be visible



## Accessing the Databases

**MySQL**

From the Host Machine:


```bash
   mysql -h 127.0.0.1 -P 3306 -u scapper_user -p

```

Basic SQL Commands:

```bash
    SHOW DATABASES;
    USE scapper_db;
    SHOW TABLES;
```

**PostgreSQL**


From the Host Machine:


```bash
    psql -h 127.0.0.1 -p 5432 -U scapper_user -d scapper_db
```

Basic SQL Commands:


```bash
    \l          -- List all databases
    \c scapper_db -- Connect to `scapper_db`
    \dt    
```


## Stopping the Application


```bash
    docker compose down
```

## Cleaning Up

```bash
    docker compose down

    docker rmi <image-id>

    docker compose up -d --build
```