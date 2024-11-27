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

Run the following commands to install aggregator into your $GOPATH

```Shell
git clone https://github.com/w0/aggregator.git
cd aggregator
go install
```

## PostgreSQL setup

On macos, assuming you followed the previous instructions to install PostgreSQL, you can run the following commands to set up the database.

Start by linking postgresql into your path. Enable the service and create the db "gator".

```Shell
brew link postgresql@15 --force
brew services start postgresql@15

createdb "gator"

```

Now that you have created the database, lets perform a migration so it is set up like `aggregator` expects.

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
		reset     resets the database. Note: this removes all data.
		users     list all registered users.
		feeds     list all available rss feeds.
		addfeed   add an rss feed to follow.
		follow    follow a feed added by a different user.
		following list feeds you are following.
		browse    list content from saved feeds.
		login     login to an existing user.
		register  register a new user.
		agg       download content from added feeds.
```
