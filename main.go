package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/jitCompileCoffee/blog-agg/internal/config"
	"github.com/jitCompileCoffee/blog-agg/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}
	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatalf("error loading database: %v", err)
	}
	dbQueries := database.New(db)
	appState := &state{
		db:  dbQueries,
		cfg: &cfg,
	}

	cmds := commands{
		cmds: make(map[string]func(*state, command) error),
	}

	cmds.register("login", handlerLogins)
	cmds.register("register", register)
	cmds.register("reset", reset)
	cmds.register("users", getUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(hanlerAddFeed))
	cmds.register("feeds", handlerListFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerGetFollows))
	cmds.register("unfollow", middlewareLoggedIn(handleUnfollow))
	cmds.register("browse", middlewareLoggedIn(handleBrowse))

	if len(os.Args) < 2 {
		log.Fatal("Usage: gator <command> [args...]")
	}

	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]

	newCmd := command{
		name: cmdName,
		args: cmdArgs,
	}

	if err := cmds.run(appState, newCmd); err != nil {
		log.Fatal(err)
	}
}
