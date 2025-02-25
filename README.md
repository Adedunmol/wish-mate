# Wish-mate
Wishlists and friendship management software.

### Features
- Manage friendships.
- Create wishlists.
- Specify when friends should get notified for wishlists.
- Get notified through mail and in-app notifications for friends' wishlists. 
- Pick items on wishlists.
- Notify friends of users on their birthdays.

### Technologies used
- [x] Golang
- [x] Goose
- [x] PostgreSQL
- [x] Chi
- [x] Redis
- [x] Docker

### Installation

#### Prerequisites:

* [Docker](https://www.docker.com/get-started)
* [Docker Compose](https://docs.docker.com/compose/install/)

##### Running the application:

1. Clone the repository
```bash
$ git clone https://github.com/Adedunmol/wish-mate.git
```

2. Change the `.env.sample` to `.env` and define the necessary environment variables.

3. Start the application in dev environment with docker compose:
```bash
$ cd wish-mate
$ docker-compose -f docker-compose.dev.yml up --build -d
```

4. Migrate the database:
```bash
$ run some script
```

5. Navigate to this endpoint `http://localhost:{PORT}/docs` to access the docs. PORT is the port defined in the `.env` file.


6. To stop the running containers, use:
```bash
$ docker-compose down
```

### Further improvements
1. Implement users picking items in fractions for the items marked as fractions.
2. Add OAuth so users can sign up with their Gmail accounts.