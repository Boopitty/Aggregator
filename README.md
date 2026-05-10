# Aggregator

## Installation

### Prerequisites

- [Go](https://golang.org/doc/install) 1.21 or later
- [PostgreSQL](https://www.postgresql.org/download/)

### Install

```bash
go install github.com/Boopitty/Aggregator
```

### Configuration
create a config file at `~/.gatorconfig.json`:

```json
{
  "db_url": "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable", "Current_User_Name":""
}
```



## Commands

### agg
You can run the aggregator in one terminal and interact with it further in a second by running:
```bash
go run . agg <Time_Interval>
```
Time interval is a time string (5s, 1m, 1h, etc).

### register
Registers a new user into the users table.
```bash
register <name>
```

### login
Login as an existing user.
```bash
login <name>
```

### users
List the names of all users in the users table.
```bash
users 
```

### reset
Removes all data from the database
```bash
reset
```

### addfeed
Adds a new feed to the feeds table. The current user will automatically follow the added feed.
```bash
addfeed <feed name> <feed URL>
```

### feeds
List all the feeds in the feeds table.
```bash
feeds
```

### follow
Follow an existing feed as the current user.
```bash
follow <feed name>
```

### following
List the names of all the feeds followed by the user
```bash
following
```

### unfollow
The current user will unfollow the given feed
```bash
unfollow <feed name>
```

### browse

Prints a certain number of posts from each followed feed. The limit defaults to 2, but can be changed.
```bash
browse <limit>
```
