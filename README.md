# aggreGator 🐊

aggreGator (🐊) is a rss feed aggregator.

## Requirements

* golang 1.23+
* postgresql 15+

On macos I installed both golang and postgressql with homebrew.

```Shell
brew install postgresql@15 go
```

## Install gator

Run the following cmds to install aggregator into your $GOPATH

```Shell
git clone https://github.com/w0/aggregator.git
cd aggregator
go install
```

## PostgreSQL setup

On macos. Assuming you installed with homebrew.

Start by link postgresql into your path, enable the service and creating the db "gator".

```Shell
brew link postgresql@15 --force
brew services start postgresql@15

createdb "gator"

```

Now that you have created the database.
Lets perform a migration so it is setup like `aggregator` expects.

```Shell
# Install Goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Perform the up migration
goose postgres "postgres://$USER:@localhost:5432/gator" up -dir sql/schema/

```

## Create .gatorconfig.json

The .gatorconfig.json should be in your home directory.

```Shell
cat << EOF > ~/.gatorconfig.json
{"db_url":"postgres://$USER:@localhost:5432/gator?sslmode=disable","current_user_name":""}
EOF
```

## Usage

### Quick Start
1. register a new user. `aggregator register $USER`
2. add a feed to watch. `aggregator addfeed TechCrunch https://techcrunch.com/feed/`
3. add content from the feed to the database. `aggregator agg 1m`
4. browse content in the database. `aggregator browse`

### Available Commands

```Shell
usage: aggregator command <arguments>
	commands:
		unfollow  stop following a feed.
		reset     resets the database. Danger this remove all data.
		users     list all registered users.
		feeds     list all available rss feeds.
		addfeed   add a rss feed to follow.
		follow    follow a feed added by a different user.
		following list feeds you are following.
		browse    list content from saved feeds.
		login     login to an existing user.
		register  register a new user.
		agg       download content from added feeds.
```
