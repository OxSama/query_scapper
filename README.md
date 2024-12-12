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